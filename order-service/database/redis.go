// order-service/database/redis.go
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheService adalah interface untuk fungsionalitas cache.
type CacheService interface {
	Get(key string) (string, error)
	SetWithTTL(key, value string, ttl time.Duration) error
	Del(keys ...string) error
}

// RedisService adalah implementasi dari CacheService.
type RedisService struct {
	client *redis.Client
}

// NewRedisService membuat dan mengembalikan sebuah RedisService.
func NewRedisService() CacheService {
	redisURI := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	client := redis.NewClient(&redis.Options{
		Addr: redisURI,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Println("Redis connection successfully opened!")
	return &RedisService{client: client}
}

// SetWithTTL mengimplementasikan method dari CacheService.
func (r *RedisService) SetWithTTL(key string, value string, ttl time.Duration) error {
	return r.client.Set(context.Background(), key, value, ttl).Err()
}

// Get mengimplementasikan method dari CacheService.
func (r *RedisService) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

// Del mengimplementasikan method dari CacheService.
func (r *RedisService) Del(keys ...string) error {
	return r.client.Del(context.Background(), keys...).Err()
}
