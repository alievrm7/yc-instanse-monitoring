package yandexapi

import (
	"net/http"
	"time"
)

type Client interface {
	ListInstancesByCloud(cloudID string) ([]Instance, error)
	ListQuotaServices() ([]string, error)
	ListQuotaLimits(cloudID, service string) ([]Quota, error)
}

type client struct {
	httpCli *http.Client
	token   string
}

func NewClient(token string) Client {
	return &client{
		httpCli: &http.Client{Timeout: 10 * time.Second},
		token:   token,
	}
}
