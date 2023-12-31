package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swafran/sclng-backend-test/internal/cache"
)

type Client struct {
	Config RepoConfig
}

type HttpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type RepoConfig struct {
	RepoUrl     string
	SearchUrl   string
	Version     string
	Token       string
	SinceOffset int
	FrontRepos  string
	Enrich      chan int
	Logger      logrus.FieldLogger
	HttpClient  HttpDoer
	RedisClient cache.GetterSetter
}

// UpdateRepos sets the latest 100 repos from Github into Redis,
// with language information
func (c *Client) UpdateRepos(ctx context.Context) {
	enrich := c.Config.Enrich
	log := c.Config.Logger
	repos := c.latest100(ctx)
	numRepos := len(repos)

	// Enrich data with languages
	for i := 0; i < numRepos; i++ {
		go c.getLanguages(ctx, &repos[i])
	}

	// Wait for all language requests to come back
	for range repos {
		<-enrich
	}

	b, err := json.Marshal(repos)
	if err != nil {
		log.Error("can't marshal repos")
	}

	err = c.Config.RedisClient.Set(c.Config.FrontRepos, string(b), 0)
	if err != nil {
		log.Error("couldn't write to redis: %s", err)
	}

	log.Info("repos updated")
}

// latest100 finds the 100 most recent repos
// first, it gets a recent repo id as a starting point,
// using the /search/repositories?q=created>=... endpoint
// then, it finds the last page of most recent repos created,
// using the /repositories?since... endpoint
func (c *Client) latest100(ctx context.Context) []Repo {
	since := c.searchSince(ctx)
	url := fmt.Sprintf("%s?since=%d", c.Config.RepoUrl, since)
	repos := []Repo{}
	c.lastCreated(ctx, url, &repos)
	if len(repos) > 100 {
		repos = repos[len(repos)-100:]
	}

	return repos
}

// searchSince finds an id among recently created repos
func (c *Client) searchSince(ctx context.Context) int {
	var since int
	log := c.Config.Logger
	recentTime := time.Now().UTC().Add(-time.Duration(c.Config.SinceOffset) * time.Minute)
	url := fmt.Sprintf("%s?q=created:>=%s", c.Config.SearchUrl, url.QueryEscape(recentTime.Format("2006-01-02T15:04:05+00:00")))

	resp, err := c.Get(ctx, url)
	if err != nil {
		log.Error(err)
	}

	searchResult := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		log.Error(err)
	}

	if len(searchResult.Items) > 0 {
		since = searchResult.Items[len(searchResult.Items)-1].Id
	}

	return since
}

// lastCreated recursively follows pages of /repositories?since=ID results
// until the last page is found, returns the repos from those results
func (c *Client) lastCreated(ctx context.Context, url string, repos *[]Repo) {
	log := c.Config.Logger

	if repos == nil {
		log.Error("can't retrieve repos")
		return
	}
	resp, err := c.Get(ctx, url)
	if err != nil {
		log.Error(err)
	}

	pageRepos := []Repo{}

	err = json.NewDecoder(resp.Body).Decode(&pageRepos)
	if err != nil {
		log.Error(err)
	}

	*repos = append(*repos, pageRepos...)
	links := c.parseLinks(resp.Header.Get("Link"))

	// Recursion
	if link, ok := links["next"]; ok {
		c.lastCreated(ctx, link, repos)
	}
}

// getLanguages follows the language link for a repo, and
// sets the language info
func (c *Client) getLanguages(ctx context.Context, repo *Repo) {
	enrich := c.Config.Enrich
	log := c.Config.Logger

	resp, err := c.Get(ctx, repo.LanguagesURL)
	if err != nil {
		log.Error(err)
	}

	languages := map[string]int{}
	err = json.NewDecoder(resp.Body).Decode(&languages)
	if err != nil {
		log.Errorf("\ncan't unmarshal resp %s: %s", repo.LanguagesURL, err)
	} else {
		for k, v := range languages {
			if repo.Languages == nil {
				repo.Languages = make(map[string]Language)
			}
			repo.Languages[k] = Language{
				Bytes: v,
			}
		}
	}

	enrich <- 1
}

// parseLinks parses "next" links and returns them as a map
func (c *Client) parseLinks(s string) map[string]string {
	links := map[string]string{}
	chunks := strings.FieldsFunc(s, func(r rune) bool {
		return r == ';' || r == ','
	})

	k, v := "", ""
	for i, chunk := range chunks {
		if i%2 == 0 {
			v = strings.Trim(chunk, "<> ")
		} else {
			k = chunk[6:len(strings.Trim(chunk, " "))]
			links[k] = v
		}
	}
	return links
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	log := c.Config.Logger
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("bad request string", err)

		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", c.Config.Version)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.Token))

	resp, err := c.Config.HttpClient.Do(req)
	if err != nil {

		// TODO set http code, error message
		log.Error("bad response", err)

		return nil, err
	}

	return resp, nil
}

// Schedule is for local scheduling only, periodically updates redis with repos
func (c *Client) Schedule(ctx context.Context, ticker *time.Ticker) {
	go func() {
		for t := range ticker.C {
			c.Config.Logger.Infof("Start repos polling %v\n", t.UTC())
			c.UpdateRepos(ctx)
		}
	}()
}
