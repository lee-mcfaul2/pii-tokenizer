package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb     redis.UniversalClient
	timeout time.Duration
}

func NewClient(cfg *config.Config) (*Client, error) {
	addrs := strings.Split(cfg.RedisAddrs, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}

	var rdb redis.UniversalClient
	switch cfg.RedisMode {
	case "single":
		if len(addrs) != 1 {
			return nil, fmt.Errorf("single mode requires exactly one address")
		}
		rdb = redis.NewClient(&redis.Options{
			Addr:     addrs[0],
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
			PoolSize: cfg.RedisPoolSize,
		})
	case "sentinel":
		if len(addrs) < 2 {
			return nil, fmt.Errorf("sentinel mode requires <master>,<sentinel1>,...")
		}
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    addrs[0],
			SentinelAddrs: addrs[1:],
			Password:      cfg.RedisPassword,
			DB:            cfg.RedisDB,
			PoolSize:      cfg.RedisPoolSize,
		})
	case "cluster":
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: cfg.RedisPassword,
			PoolSize: cfg.RedisPoolSize,
		})
	default:
		return nil, fmt.Errorf("unknown redis mode: %s", cfg.RedisMode)
	}

	return &Client{rdb: rdb, timeout: cfg.RedisTimeout}, nil
}

func (c *Client) Do(ctx context.Context, fn func(ctx context.Context, r redis.UniversalClient) error) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return fn(ctx, c.rdb)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Do(ctx, func(ctx context.Context, r redis.UniversalClient) error {
		return r.Ping(ctx).Err()
	})
}

func (c *Client) Close() error { return c.rdb.Close() }
