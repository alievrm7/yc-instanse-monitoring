package yandexapi

import (
	"context"
	"encoding/json"
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

// приватный метод для чтения токена
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

// универсальная функция для GET-запросов (дженерик)
func apiGet[T any](ctx context.Context, cli *http.Client, token string, url string, out *T) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s failed: status=%d", url, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s failed: %w", url, err)
	}
	return nil
}
