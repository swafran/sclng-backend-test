package cache

import (
	"time"

	"github.com/go-redis/redis"
)

type GetterSetter interface {
	Get(string) (string, error)
	Set(string, interface{}, time.Duration) error
}

type client struct {
	client *redis.Client
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func StartClient(cfg RedisConfig) (client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		return client{}, err
	}

	return client{
		client: redisClient,
	}, err
}

func (c *client) Get(s string) (string, error) {
	stringCmd := c.client.Get(s)

	return stringCmd.Val(), stringCmd.Err()
}

func (c *client) Set(k string, v interface{}, expiration time.Duration) error {
	statusCmd := c.client.Set(k, v, expiration)

	return statusCmd.Err()
}
