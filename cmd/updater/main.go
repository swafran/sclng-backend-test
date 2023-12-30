package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/swafran/sclng-backend-test/internal/cache"
	gh "github.com/swafran/sclng-backend-test/internal/github"
)

func main() {
	log := logger.Default()
	log.Info("Initializing app")
	cfg, err := newConfig()
	if err != nil {
		log.WithError(err).Error("failed to initialize configuration")
		os.Exit(1)
	}

	redisClient, err := cache.StartClient(cache.RedisConfig{
		Addr:     cfg.RedisUrl,
		Password: "",
		DB:       0,
	})
	if err != nil {
		log.WithError(err).Error("failed to establish redis connection")
		os.Exit(1)
	}

	repoConfig := gh.RepoConfig{
		RepoUrl:     fmt.Sprintf("%s%s", cfg.GithubApi, cfg.GithubRepos),
		SearchUrl:   fmt.Sprintf("%s%s", cfg.GithubApi, cfg.GithubSearch),
		Version:     cfg.GithubVersion,
		Token:       cfg.GithubToken,
		SinceOffset: cfg.SinceOffset,
		FrontRepos:  cfg.FrontRepos,
		Enrich:      make(chan int),
		Logger:      log,
		HttpClient:  cfg.Services.HttpClient,
		RedisClient: &redisClient,
	}
	githubClient := gh.Client{
		Config: repoConfig,
	}
	cfg.Services.GithubClient = githubClient

	//TODO
	ctx := context.Background()
	githubClient.UpdateRepos(ctx)

	if cfg.LocalSchedule {
		forever := make(chan bool)
		ticker := time.NewTicker(time.Duration(cfg.LocalSchedulePeriod) * time.Minute)
		defer ticker.Stop()

		log.Info("starting local scheduler")
		githubClient.Schedule(ctx, ticker)

		<-forever
	}
}
