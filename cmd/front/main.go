package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/go-redis/redis"
)

func main() {
	log := logger.Default()
	log.Info("Initializing app")
	cfg, err := newConfig()
	if err != nil {
		log.WithError(err).Error("Fail to initialize configuration")
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisUrl,
		Password: "",
		DB:       0,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		// TODO
		log.Error(err)
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", pongHandler)
	// Initialize web server and configure the following routes:
	// GET /repos
	// GET /stats

	router.HandleFunc("/repos", reposHandler)

	log = log.WithField("port", cfg.Port)
	log.Info("Listening...")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
	if err != nil {
		log.WithError(err).Error("Fail to listen to the given port")
		os.Exit(2)
	}
}

func reposHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) error {

}

func pongHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) error {
	log := logger.Get(r.Context())
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
	if err != nil {
		log.WithError(err).Error("Fail to encode JSON")
	}
	return nil
}
