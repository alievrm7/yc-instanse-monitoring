package yandexapi

import (
	"net/http"
	"time"
)

type Client interface {
	ListInstances(folderID string) ([]Instance, error)
	ListClouds() ([]Cloud, error)
	ListFolders(cloudID string) ([]Folder, error)
	ListAllInstances() ([]Instance, error)
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
