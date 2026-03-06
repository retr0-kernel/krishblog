package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(redisURL string) (*Redis, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	opts.PoolSize = 10
	opts.MinIdleConns = 2
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.Client.Close()
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.Client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

func (r *Redis) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := r.Client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("redis pipeline: %w", err)
	}
	return incr.Val(), nil
}

func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.RPush(ctx, key, values...).Err()
}

func (r *Redis) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.LRange(ctx, key, start, stop).Result()
}

func (r *Redis) LTrim(ctx context.Context, key string, start, stop int64) error {
	return r.Client.LTrim(ctx, key, start, stop).Err()
}
