package redis

import (
	"LAB1/internal/app/config"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "jwt_blacklist:" // наш префикс

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := &Client{}
	client.cfg = cfg

	redisClient := redis.NewClient(&redis.Options{
		Password:    cfg.Password,
		Username:    cfg.User,
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// WriteJWTToBlacklist записывает JWT в блеклист
func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, ttl time.Duration, userInfo string) error {
	key := servicePrefix + jwtStr
	return c.client.Set(ctx, key, userInfo, ttl).Err()
}

// CheckJWTInBlacklist проверяет наличие JWT в блеклисте
func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	key := servicePrefix + jwtStr
	return c.client.Get(ctx, key).Err()
}

// GetAllJWTKeys возвращает все ключи JWT из blacklist (для демонстрации)
func (c *Client) GetAllJWTKeys(ctx context.Context) ([]string, error) {
	pattern := servicePrefix + "*"
	return c.client.Keys(ctx, pattern).Result()
}

// GetJWTInfo возвращает информацию о JWT (для демонстрации)
func (c *Client) GetJWTInfo(ctx context.Context, jwtStr string) (string, error) {
	key := servicePrefix + jwtStr
	return c.client.Get(ctx, key).Result()
}
