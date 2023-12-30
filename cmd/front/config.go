package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	Port     int    `envconfig:"PORT" default:"5000"`
	RedisUrl string `envconfig:"REDIS_URL" default:"redis:6379"`
}

func newConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build config from env")
	}
	return &cfg, nil
}
