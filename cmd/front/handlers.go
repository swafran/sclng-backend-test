package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/swafran/sclng-backend-test/internal/cache"
	gh "github.com/swafran/sclng-backend-test/internal/github"
)

func pongHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) error {
	log := logger.Get(r.Context())
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
	if err != nil {
		log.WithError(err).Error("Fail to encode JSON")
		return err
	}

	return nil
}

// reposHandler writes the 100 latest repositories data, returns error
func reposHandler(c cache.GetterSetter, w http.ResponseWriter, r *http.Request) error {
	log := logger.Get(r.Context())
	reposJson := ""
	params := r.URL.Query()
	fullPath := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)

	// if there are query parameters, first try with full path as key
	if len(params) > 0 {
		cachedQuery, err := c.Get(fullPath)
		if err != nil {
			log.WithError(err).Warn("error retrieving full path from cache")
		} else if cachedQuery != "" {
			reposJson = cachedQuery
		}
	}

	// then try with route path as key
	// (this was cached by the updater)
	if reposJson == "" {
		reposJson, err := c.Get("/repos")
		if err != nil {
			log.WithError(err).Error("error retrieving default path from cache")
		}

		// apply filter as necessary
		if reqLangs, ok := params["l"]; ok {
			reposJson, err = filterByLanguages(reposJson, reqLangs)

			if err != nil {
				log.WithError(err).Error("can't apply language filter")
			} else {

				// cache the filtered result by full path (including query params)
				err = c.Set(fullPath, reposJson, 0)
				if err != nil {
					log.WithError(err).Error("can't cache filtered results")
				}
			}
		}
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(reposJson))

	return nil
}

// statsHandler writes stats on the api, returns error
func statsHandler(c cache.GetterSetter, w http.ResponseWriter, r *http.Request) error {
	json, err := c.Get("/stats")
	if err != nil {
		//TODO set status & write
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(json))

	return nil
}

// cacheHandler takes a cache client, and a function with a signature like f() below
// returns that function as a handlers.HandlerFunc for adding to handlers.Router
func cacheHandler(c cache.GetterSetter,
	f func(c cache.GetterSetter, w http.ResponseWriter, r *http.Request) error) handlers.HandlerFunc {
	return handlers.HandlerFunc(func(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
		return f(c, w, r)
	})
}

func filterByLanguages(reposJson string, reqLangs []string) (string, error) {
	acceptedLangs := map[string]struct{}{}
	for _, l := range reqLangs {
		acceptedLangs[l] = struct{}{}
	}
	repos := []gh.Repo{}
	err := json.Unmarshal([]byte(reposJson), &repos)
	if err != nil {
		return reposJson, err
	}

	filteredRepos := []gh.Repo{}
	for _, repo := range repos {
		if repo.Languages != nil {
			for language := range repo.Languages {
				if _, ok := acceptedLangs[language]; ok {
					filteredRepos = append(filteredRepos, repo)
				}
			}
		}
	}

	b, err := json.Marshal(filteredRepos)
	if err != nil {
		return reposJson, err
	}
	return string(b), nil
}
