package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"api/internal/config"
	"api/internal/database"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(redisConfig *config.RedisConfig) (*Client, error) {
	options := &redis.Options{
		Addr:     redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	}
	if redisConfig.TLS {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: redisConfig.Host,
		}
		if redisConfig.TLSCA != "" {
			rootCAs, err := database.CertificatePool(redisConfig.TLSCA)
			if err != nil {
				return nil, fmt.Errorf("configure Redis TLS CA: %w", err)
			}
			options.TLSConfig.RootCAs = rootCAs
		}
	} else if redisConfig.TLSCA != "" {
		return nil, fmt.Errorf("REDIS_TLS_CA requires REDIS_TLS=true")
	}
	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}

	return &Client{client: client}, nil
}

func (r *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}
	return nil
}

func (r *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis get %s: %w", key, err)
	}
	return val, nil
}

func (r *Client) Del(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del %s: %w", key, err)
	}
	return nil
}

func (r *Client) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("close redis client: %w", err)
	}
	return nil
}
