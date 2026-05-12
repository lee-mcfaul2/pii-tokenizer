//go:build integration

package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func newRedisClient(t *testing.T) *Client {
	t.Helper()
	ctx := context.Background()
	c, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Terminate(ctx) })
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "6379")
	client, err := NewClient(&config.Config{
		RedisMode:    "single",
		RedisAddrs:   host + ":" + port.Port(),
		RedisTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestScope_InitGetRelease(t *testing.T) {
	c := newRedisClient(t)
	ctx := context.Background()
	now := time.Now()

	e := ScopeEntry{
		WrappedKReq:    []byte("fake-wrapped"),
		KMasterVersion: 1,
		CreatedAt:      now,
		ExpiresAt:      now.Add(60 * time.Second),
	}
	if err := c.Init(ctx, "uuid-a", e, 60*time.Second); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, "uuid-a")
	if err != nil {
		t.Fatal(err)
	}
	if got.KMasterVersion != 1 {
		t.Errorf("version: got %d", got.KMasterVersion)
	}
	if string(got.WrappedKReq) != "fake-wrapped" {
		t.Errorf("payload mismatch")
	}
	if err := c.Release(ctx, "uuid-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "uuid-a"); !errors.Is(err, ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound, got %v", err)
	}
}

func TestScope_InitIdempotent(t *testing.T) {
	c := newRedisClient(t)
	ctx := context.Background()
	now := time.Now()

	e := ScopeEntry{
		WrappedKReq:    []byte("v1"),
		KMasterVersion: 1,
		CreatedAt:      now,
		ExpiresAt:      now.Add(60 * time.Second),
	}
	if err := c.Init(ctx, "uuid-x", e, 60*time.Second); err != nil {
		t.Fatal(err)
	}
	e2 := e
	e2.WrappedKReq = []byte("v2")
	err := c.Init(ctx, "uuid-x", e2, 60*time.Second)
	var ae *ErrScopeAlreadyExists
	if !errors.As(err, &ae) {
		t.Fatalf("expected ErrScopeAlreadyExists, got %v", err)
	}
	got, _ := c.Get(ctx, "uuid-x")
	if string(got.WrappedKReq) != "v1" {
		t.Errorf("second init overwrote entry")
	}
}

func TestScope_GetMissing(t *testing.T) {
	c := newRedisClient(t)
	if _, err := c.Get(context.Background(), "no-such-uuid"); !errors.Is(err, ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound, got %v", err)
	}
}
