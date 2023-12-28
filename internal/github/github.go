package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

func Search100(ctx context.Context, searchUrl string, log logrus.FieldLogger, httpclient *http.Client) {

	query := queryString()
	searchUrl = fmt.Sprintf("%s?q=%s&sort=created", searchUrl, query)

	fmt.Println(searchUrl)
	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		log.Error("bad request string", err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {

		// TODO set http code, error message
		log.Error("bad response", err)
	}

	result := SearchResult{}
	err2 := json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println(err2)

	fmt.Println(len(result.Items))
	// fmt.Println(repos[0])

	for _, repo := range result.Items {
		fmt.Println(repo.Id)
	}

}

func queryString() string {
	return url.QueryEscape("created:>=2023-12-26")
}
