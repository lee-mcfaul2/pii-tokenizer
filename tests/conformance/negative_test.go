//go:build integration

package conformance

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

func bootForConformance(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()
	rc, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rc.Terminate(ctx) })
	host, _ := rc.Host(ctx)
	port, _ := rc.MappedPort(ctx, "6379")

	cli, _ := store.NewClient(&config.Config{
		RedisMode: "single", RedisAddrs: host + ":" + port.Port(),
		RedisTimeout: 2 * time.Second,
	})
	t.Cleanup(func() { cli.Close() })

	keyDir := t.TempDir()
	k := make([]byte, 32)
	rand.Read(k)
	os.WriteFile(filepath.Join(keyDir, "k_master_v1"), k, 0o600)
	os.WriteFile(filepath.Join(keyDir, "current"), []byte("v1"), 0o600)
	km, _ := kmaster.NewAEADLocal(ctx, &config.Config{AEADKeyPath: keyDir})
	t.Cleanup(func() { km.Close() })

	svc := scope.New(cli, km)
	srv := api.NewServer(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	srv.Ready = func(ctx context.Context) error { return cli.Ping(ctx) }

	ts := httptest.NewServer(srv.Router())
	t.Cleanup(ts.Close)

	body, _ := json.Marshal(map[string]any{
		"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "ttl_seconds": 60,
	})
	resp, _ := http.Post(ts.URL+"/v1/init_request", "application/json", bytes.NewReader(body))
	resp.Body.Close()
	return ts
}

func TestConformance_NegativeCases(t *testing.T) {
	ts := bootForConformance(t)

	for _, c := range Cases {
		t.Run(c.Name, func(t *testing.T) {
			body, _ := json.Marshal(c.RequestBody)
			req, _ := http.NewRequest("POST", ts.URL+c.Endpoint, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != c.WantHTTPStatus {
				t.Errorf("status: got %d want %d", resp.StatusCode, c.WantHTTPStatus)
			}
			var env api.ErrorEnvelope
			if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
				t.Fatalf("decode envelope: %v", err)
			}
			if env.ErrorType != c.WantErrorType {
				t.Errorf("error_type: got %s want %s", env.ErrorType, c.WantErrorType)
			}
		})
	}
}
