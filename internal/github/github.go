package github

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

func Search100(ctx context.Context, searchUrl string, log logrus.FieldLogger, httpclient *http.Client) {

	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		log.Error("bad request string", err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Error("bad request string", err)
	}

}
