package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(redisURL string) error {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis url: %v", err)
	}

	// Upstash requires TLS if using rediss://
	if opt.TLSConfig == nil && len(opt.Password) > 0 {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	Client = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping redis: %v", err)
	}

	fmt.Println("redis connection successful")
	return nil
}
