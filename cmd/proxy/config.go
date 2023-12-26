package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	Port         int    `envconfig:"PORT" default:"5000"`
	GithubApi    string `envconfig:"GITHUB_API" default:"https://api.github.com"`
	GithubSearch string
}

func newConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("sbt", &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build config from env")
	}
	return &cfg, nil
}
