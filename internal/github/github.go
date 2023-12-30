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

func (c *Client) UpdateRepos(ctx context.Context) {
	enrich := c.Config.Enrich

	log := c.Config.Logger

	repos := c.latest100(ctx)

	fmt.Println(len(repos))

	numRepos := len(repos)

	for i := 0; i < numRepos; i++ {
		go c.getLanguages(ctx, &repos[i])
	}

	for range repos {
		<-enrich
	}

	b, err := json.Marshal(repos)
	if err != nil {
		// TODO
	}

	err = c.Config.RedisClient.Set(c.Config.FrontRepos, string(b), 0)
	if err != nil {
		log.Error("couldn't write to redis: %s", err)
	}

	log.Info("wrote to redis")
	fmt.Println("done!")
}

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

func (c *Client) searchSince(ctx context.Context) int {
	var since int
	log := c.Config.Logger
	recentTime := time.Now().UTC().Add(-time.Duration(c.Config.SinceOffset) * time.Minute)
	url := fmt.Sprintf("%s?q=created:>=%s", c.Config.SearchUrl, url.QueryEscape(recentTime.Format("2006-01-02T15:04:05+00:00")))

	resp, err := c.Get(ctx, url)
	if err != nil {
		// TODO
		log.Error("really bad ", err)
	}

	searchResult := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		// TODO
		fmt.Println(err)
	}

	if len(searchResult.Items) > 0 {
		since = searchResult.Items[len(searchResult.Items)-1].Id
	}
	// TODO else error

	return since
}

func (c *Client) lastCreated(ctx context.Context, url string, repos *[]Repo) {
	fmt.Printf("\nstarting lastest100 url:%s, repos count:%d ", url, len(*repos))

	if repos == nil {
		// TODO error
		return
	}

	log := c.Config.Logger
	resp, err := c.Get(ctx, url)
	// TODO error

	pageRepos := []Repo{}

	err = json.NewDecoder(resp.Body).Decode(&pageRepos)
	if err != nil {
		// TODO
		log.Error("bad XXXXX")
	}

	fmt.Println("pageRepos: ", pageRepos)
	*repos = append(*repos, pageRepos...)

	links := c.parseLinks(resp.Header.Get("Link"))

	// Recursion
	if link, ok := links["next"]; ok {
		c.lastCreated(ctx, link, repos)
	}
}

func (c *Client) getLanguages(ctx context.Context, repo *Repo) {
	enrich := c.Config.Enrich
	log := c.Config.Logger

	fmt.Println("processing: ", repo.Id)

	resp, err := c.Get(ctx, repo.LanguagesURL)
	if err != nil {
		// TODO
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

func (c *Client) Schedule(ctx context.Context, ticker *time.Ticker) {
	go func() {
		for t := range ticker.C {
			c.Config.Logger.Infof("Start repos polling %v\n", t.UTC())
			c.UpdateRepos(ctx)
		}
	}()
}
