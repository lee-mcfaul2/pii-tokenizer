//go:build integration

package store

import (
	"context"
	"testing"
	"time"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestClient_Ping(t *testing.T) {
	ctx := context.Background()
	c, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Terminate(ctx)
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "6379")

	client, err := NewClient(&config.Config{
		RedisMode:    "single",
		RedisAddrs:   host + ":" + port.Port(),
		RedisTimeout: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
