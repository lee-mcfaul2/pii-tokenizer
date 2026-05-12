//go:build integration

package scope

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/kmaster"
	"github.com/lee-mcfaul2/pii-tokenizer/internal/store"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setup(t *testing.T) *Service {
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
		RedisMode:    "single",
		RedisAddrs:   host + ":" + port.Port(),
		RedisTimeout: 2 * time.Second,
	})
	t.Cleanup(func() { cli.Close() })

	keyDir := t.TempDir()
	k1 := make([]byte, 32)
	rand.Read(k1)
	os.WriteFile(filepath.Join(keyDir, "k_master_v1"), k1, 0o600)
	os.WriteFile(filepath.Join(keyDir, "current"), []byte("v1"), 0o600)
	km, _ := kmaster.NewAEADLocal(ctx, &config.Config{AEADKeyPath: keyDir})
	t.Cleanup(func() { km.Close() })

	return New(cli, km)
}

func TestScope_HappyPath(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	uuid := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	res, err := s.InitRequest(ctx, uuid, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Error("expected Created=true on first init")
	}

	tok, err := s.Tokenize(ctx, uuid, "EMAIL", []byte("alice@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	piiType, pt, err := s.Detokenize(ctx, uuid, tok)
	if err != nil {
		t.Fatal(err)
	}
	if piiType != "EMAIL" || string(pt) != "alice@example.com" {
		t.Errorf("roundtrip: got (%s, %q)", piiType, pt)
	}

	if err := s.ReleaseRequest(ctx, uuid); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Tokenize(ctx, uuid, "EMAIL", []byte("x")); !errors.Is(err, store.ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound after release, got %v", err)
	}
}

func TestScope_IdempotentInit(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	uuid := "uuid-idempotent"

	r1, _ := s.InitRequest(ctx, uuid, 30*time.Second)
	r2, _ := s.InitRequest(ctx, uuid, 30*time.Second)

	if !r1.Created || r2.Created {
		t.Errorf("Created flags wrong: r1=%v r2=%v", r1.Created, r2.Created)
	}
	if !r1.ExpiresAt.Equal(r2.ExpiresAt) {
		t.Errorf("expires_at should be unchanged on idempotent init")
	}
}

func TestScope_CrossRequestIsolation(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	s.InitRequest(ctx, "uuid-a", 30*time.Second)
	s.InitRequest(ctx, "uuid-b", 30*time.Second)

	tok1, _ := s.Tokenize(ctx, "uuid-a", "EMAIL", []byte("alice@example.com"))
	tok2, _ := s.Tokenize(ctx, "uuid-b", "EMAIL", []byte("alice@example.com"))

	if tok1 == tok2 {
		t.Error("same plaintext across requests must produce different tokens")
	}

	if _, _, err := s.Detokenize(ctx, "uuid-b", tok1); err == nil {
		t.Error("expected detokenize to fail when crossing requests")
	}
}
