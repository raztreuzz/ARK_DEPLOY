package redis

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return errors.New("REDIS_URL is required")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	Client = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

func CloseRedis() error {
	if Client == nil {
		return nil
	}
	return Client.Close()
}