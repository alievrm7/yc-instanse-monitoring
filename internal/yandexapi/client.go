package yandexapi

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client interface {
	ListInstancesByCloud(cloudID string) ([]Instance, error)
	ListQuotaServices() ([]string, error)
	ListQuotaLimits(cloudID, service string) ([]Quota, error)
	ListFolders(cloudID string) ([]Folder, error)
}

type client struct {
	httpCli   *http.Client
	tokenFile string
}

func NewClient(tokenFile string) Client {
	return &client{
		httpCli:   &http.Client{Timeout: 10 * time.Second},
		tokenFile: tokenFile,
	}
}

// приватный метод, чтобы читать токен каждый раз
func (c *client) getToken() (string, error) {
	b, err := os.ReadFile(c.tokenFile)
	if err != nil {
		return "", fmt.Errorf("read token file: %w", err)
	}
	tok := strings.TrimSpace(string(b))
	if tok == "" {
		return "", errors.New("token file is empty")
	}
	return tok, nil
}
