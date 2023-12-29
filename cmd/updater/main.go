package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Scalingo/go-utils/logger"
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

	ctx := context.Background()

	repoConfig := gh.RepoConfig{
		RepoUrl:    fmt.Sprintf("%s%s", cfg.GithubApi, cfg.GithubRepos),
		LangUrl:    fmt.Sprintf("%s%s", cfg.GithubApi, cfg.GithubRepos),
		Enrich:     make(chan int),
		Logger:     log,
		HttpClient: cfg.Services.HttpClient,
	}
	githubClient := gh.Client{
		Config: repoConfig,
	}
	cfg.Services.GithubClient = githubClient

	githubClient.UpdateRepos(ctx)

	if cfg.LocalSchedule {
		log.Info("starting local schedule")
		githubClient.Dispatch()
	}
}
