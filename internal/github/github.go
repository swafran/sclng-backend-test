package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Client struct {
	Config RepoConfig
}

type RepoConfig struct {
	RepoUrl    string
	LangUrl    string
	Enrich     chan int
	Logger     logrus.FieldLogger
	HttpClient *http.Client
}

func (c *Client) UpdateRepos(ctx context.Context) {
	enrich := c.Config.Enrich
	url := fmt.Sprintf("%s?since=730000000", c.Config.RepoUrl)
	log := c.Config.Logger

	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("bad request string", err)
	}

	resp, err := c.Config.HttpClient.Do(req)
	if err != nil {

		// TODO set http code, error message
		log.Error("bad response", err)
	}

	repos := []Repo{}
	err2 := json.NewDecoder(resp.Body).Decode(&repos)

	fmt.Println(err2)

	fmt.Println(len(repos))

	numRepos := len(repos)

	for i := 0; i < numRepos; i++ {
		go c.getLanguages(ctx, &repos[i])
	}

	for range repos {
		<-enrich
	}

	fmt.Println("done")
}

func (c *Client) getLanguages(ctx context.Context, repo *Repo) {
	enrich := c.Config.Enrich
	log := c.Config.Logger

	fmt.Println("processing: ", repo.Id)

	req, err := http.NewRequest("GET", repo.LanguagesURL, nil)
	if err != nil {
		log.Error("bad request string", err)
	}

	resp, err := c.Config.HttpClient.Do(req)
	if err != nil {

		// TODO set http code, error message
		log.Error("bad response", err)
	}

	languages := map[string]int{}
	err = json.NewDecoder(resp.Body).Decode(&languages)
	if err != nil {
		log.Error("url: " + repo.LanguagesURL)
		log.Error("can't unmarshal response ", err)
	}

	for k, v := range languages {
		repo.Languages[k] = Language{
			Bytes: v,
		}
	}

	enrich <- 1
}

func (c *Client) Dispatch() {

}
