package main

import (
	"context"
	"fmt"
	"net/http"
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

	// httpclient := &http.Client{Timeout: 10 * time.Second}

	httpclient := &http.Client{}
	ctx := context.Background()

	searchUrl := fmt.Sprintf("%s%s", cfg.GithubApi, cfg.GithubSearch)
	gh.Search100(ctx, searchUrl, log, httpclient)
}
