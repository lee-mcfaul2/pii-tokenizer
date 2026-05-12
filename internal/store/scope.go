package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "request:"

type ScopeEntry struct {
	WrappedKReq    []byte
	KMasterVersion int
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

var ErrScopeNotFound = errors.New("scope not found")
var ErrScopeExpired = errors.New("scope expired")

type ErrScopeAlreadyExists struct {
	ExpiresAt time.Time
}

func (e *ErrScopeAlreadyExists) Error() string {
	return fmt.Sprintf("scope already exists; expires_at=%s", e.ExpiresAt.UTC().Format(time.RFC3339))
}

const initLua = `
if redis.call("EXISTS", KEYS[1]) == 1 then
  return 0
end
redis.call("HSET", KEYS[1],
  "wrapped_kreq",     ARGV[1],
  "k_master_version", ARGV[2],
  "created_at",       ARGV[3],
  "exp",              ARGV[4])
redis.call("EXPIRE", KEYS[1], ARGV[5])
return 1
`

func (c *Client) Init(ctx context.Context, uuid string, e ScopeEntry, ttl time.Duration) error {
	key := keyPrefix + uuid
	script := redis.NewScript(initLua)

	return c.Do(ctx, func(ctx context.Context, r redis.UniversalClient) error {
		res, err := script.Run(ctx, r, []string{key},
			e.WrappedKReq,
			fmt.Sprintf("%d", e.KMasterVersion),
			e.CreatedAt.UTC().Format(time.RFC3339),
			e.ExpiresAt.UTC().Format(time.RFC3339),
			int(ttl.Seconds()),
		).Int()
		if err != nil {
			return err
		}
		if res == 0 {
			existing, err := c.getRaw(ctx, r, uuid)
			if err != nil {
				return err
			}
			return &ErrScopeAlreadyExists{ExpiresAt: existing.ExpiresAt}
		}
		return nil
	})
}

func (c *Client) Get(ctx context.Context, uuid string) (*ScopeEntry, error) {
	var entry *ScopeEntry
	err := c.Do(ctx, func(ctx context.Context, r redis.UniversalClient) error {
		e, err := c.getRaw(ctx, r, uuid)
		entry = e
		return err
	})
	return entry, err
}

func (c *Client) Release(ctx context.Context, uuid string) error {
	return c.Do(ctx, func(ctx context.Context, r redis.UniversalClient) error {
		return r.Del(ctx, keyPrefix+uuid).Err()
	})
}

func (c *Client) getRaw(ctx context.Context, r redis.UniversalClient, uuid string) (*ScopeEntry, error) {
	key := keyPrefix + uuid
	fields, err := r.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, ErrScopeNotFound
	}
	exp, err := time.Parse(time.RFC3339, fields["exp"])
	if err != nil {
		return nil, fmt.Errorf("parse exp: %w", err)
	}
	if time.Now().After(exp) {
		return nil, ErrScopeExpired
	}
	created, _ := time.Parse(time.RFC3339, fields["created_at"])
	var version int
	fmt.Sscanf(fields["k_master_version"], "%d", &version)

	return &ScopeEntry{
		WrappedKReq:    []byte(fields["wrapped_kreq"]),
		KMasterVersion: version,
		CreatedAt:      created,
		ExpiresAt:      exp,
	}, nil
}
