package redis

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Error parsing Redis URL: %v", err)
		return err
	}

	Client = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*1000000000)
	defer cancel()

	_, err = Client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
		return err
	}

	log.Println("✓ Redis connected successfully")
	return nil
}

func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
