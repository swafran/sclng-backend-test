package main

import (
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	gh "github.com/swafran/sclng-backend-test/internal/github"
)

type Config struct {
	Port                int    `envconfig:"PORT" default:"5001"`
	GithubApi           string `envconfig:"GITHUB_API" default:"https://api.github.com"`
	GithubRepos         string `envconfig:"GITHUB_REPOS" default:"/repositories"`
	GithubSearch        string `envconfig:"GITHUB_SEARCH" default:"/search/repositories"`
	GithubVersion       string `envconfig:"GITHUB_VERSION" default:"2022-11-28"`
	GithubToken         string `envconfig:"GITHUB_TOKEN"`
	SinceOffset         int    `envconfig:"SINCE_OFFSET"`
	LocalSchedule       bool   `envconfig:"LOCAL_SCHEDULE" default:"false"`
	LocalSchedulePeriod int    `envconfig:"LOCAL_SCHEDULE_PERIOD" default:"10"`
	RedisUrl            string `envconfig:"REDIS_URL"`
	FrontRepos          string `envconfig:"FRONT_REPOS"`
	Services            Services
}

type Services struct {
	Logger       logrus.FieldLogger
	HttpClient   *http.Client
	GithubClient gh.Client
}

func newConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("sbt", &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build config from env")
	}
	cfg.Services = Services{
		HttpClient: &http.Client{},
	}
	return &cfg, nil
}
