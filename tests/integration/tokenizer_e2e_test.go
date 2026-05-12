//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/api"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/kmaster"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/scope"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/store"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func bootTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()

	rc, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rc.Terminate(ctx) })
	host, _ := rc.Host(ctx)
	port, _ := rc.MappedPort(ctx, "6379")

	cli, err := store.NewClient(&config.Config{
		RedisMode:    "single",
		RedisAddrs:   host + ":" + port.Port(),
		RedisTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cli.Close() })

	keyDir := t.TempDir()
	k := make([]byte, 32)
	rand.Read(k)
	os.WriteFile(filepath.Join(keyDir, "k_master_v1"), k, 0o600)
	os.WriteFile(filepath.Join(keyDir, "current"), []byte("v1"), 0o600)
	km, err := kmaster.NewAEADLocal(ctx, &config.Config{AEADKeyPath: keyDir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { km.Close() })

	svc := scope.New(cli, km)
	srv := api.NewServer(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	srv.Ready = func(ctx context.Context) error { return cli.Ping(ctx) }

	ts := httptest.NewServer(srv.Router())
	t.Cleanup(ts.Close)
	return ts
}

func postJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

const fixtureUUID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"

func TestE2E_FullRoundTrip(t *testing.T) {
	ts := bootTestServer(t)

	resp := postJSON(t, ts.URL+"/v1/init_request", map[string]any{
		"request_uuid": fixtureUUID, "ttl_seconds": 60,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("init: status %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = postJSON(t, ts.URL+"/v1/tokenize", map[string]any{
		"request_uuid": fixtureUUID, "type": "EMAIL", "plaintext": "alice@example.com",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("tokenize: status %d", resp.StatusCode)
	}
	var tr struct{ Token string }
	json.NewDecoder(resp.Body).Decode(&tr)
	resp.Body.Close()
	if tr.Token == "" {
		t.Fatal("empty token")
	}

	resp = postJSON(t, ts.URL+"/v1/detokenize", map[string]any{
		"request_uuid": fixtureUUID, "token": tr.Token,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("detokenize: status %d", resp.StatusCode)
	}
	var dr struct {
		Plaintext string `json:"plaintext"`
		Type      string `json:"type"`
	}
	json.NewDecoder(resp.Body).Decode(&dr)
	resp.Body.Close()
	if dr.Plaintext != "alice@example.com" || dr.Type != "EMAIL" {
		t.Errorf("roundtrip: %+v", dr)
	}

	resp = postJSON(t, ts.URL+"/v1/release_request", map[string]any{"request_uuid": fixtureUUID})
	if resp.StatusCode != 204 {
		t.Fatalf("release: status %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = postJSON(t, ts.URL+"/v1/tokenize", map[string]any{
		"request_uuid": fixtureUUID, "type": "EMAIL", "plaintext": "x",
	})
	if resp.StatusCode != 404 {
		t.Fatalf("post-release status: %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_DeterminismWithinRequest(t *testing.T) {
	ts := bootTestServer(t)
	resp := postJSON(t, ts.URL+"/v1/init_request", map[string]any{
		"request_uuid": fixtureUUID, "ttl_seconds": 60,
	})
	resp.Body.Close()

	get := func() string {
		r := postJSON(t, ts.URL+"/v1/tokenize", map[string]any{
			"request_uuid": fixtureUUID, "type": "EMAIL", "plaintext": "x@y",
		})
		var tr struct{ Token string }
		json.NewDecoder(r.Body).Decode(&tr)
		r.Body.Close()
		return tr.Token
	}
	a, b := get(), get()
	if a != b {
		t.Errorf("non-deterministic: %s vs %s", a, b)
	}
}

func TestE2E_TamperedTokenRejected(t *testing.T) {
	ts := bootTestServer(t)
	postJSON(t, ts.URL+"/v1/init_request", map[string]any{
		"request_uuid": fixtureUUID, "ttl_seconds": 60,
	}).Body.Close()

	r := postJSON(t, ts.URL+"/v1/tokenize", map[string]any{
		"request_uuid": fixtureUUID, "type": "EMAIL", "plaintext": "alice@example.com",
	})
	var tr struct{ Token string }
	json.NewDecoder(r.Body).Decode(&tr)
	r.Body.Close()

	tampered := []byte(tr.Token)
	tampered[len(tampered)-1] ^= 0x01

	resp := postJSON(t, ts.URL+"/v1/detokenize", map[string]any{
		"request_uuid": fixtureUUID, "token": string(tampered),
	})
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 AAD_MISMATCH, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_HealthAndReady(t *testing.T) {
	ts := bootTestServer(t)
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("healthz: %v %d", err, resp.StatusCode)
	}
	resp.Body.Close()
	resp, err = http.Get(ts.URL + "/readyz")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("readyz: %v %d", err, resp.StatusCode)
	}
	resp.Body.Close()
}
