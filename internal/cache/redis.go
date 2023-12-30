package cache

import "github.com/go-redis/redis"

type Cacher interface {
	Set(string, string)
	Get(string) string
}

type client struct {
	client redis.Client
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
		redisClient,
	}
}

func (c *client) Get(s string) string {
	return c.client.Get(s)
}

func (c *client) Set(k string, v string) {
	statusCmd := c.Client.Set(k, v, 0)

	// treat statusCmd...
}
