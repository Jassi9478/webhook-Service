package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyStore struct {
	client *redis.Client
}

func NewIdempotencyStore(addr string) *IdempotencyStore {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &IdempotencyStore{client: rdb}
}

func (s *IdempotencyStore) Lock(ctx context.Context, key string) (bool, error) {
	// 24-hour expiration on idempotency keys
	return s.client.SetNX(ctx, "idempotency:"+key, "locked", 24*time.Hour).Result()
}
