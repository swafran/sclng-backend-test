package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/swafran/sclng-backend-test/internal/cache"
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
		os.Exit(2)
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", pongHandler)
	router.HandleFunc("/repos", cacheHandler(&redisClient, reposHandler))
	router.HandleFunc("/stats", cacheHandler(&redisClient, statsHandler))

	log = log.WithField("port", cfg.Port)
	log.Info("Listening...")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
	if err != nil {
		log.WithError(err).Error("failed to listen to the given port")
		os.Exit(3)
	}
}
