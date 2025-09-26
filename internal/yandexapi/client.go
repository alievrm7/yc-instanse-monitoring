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

// generic helper с пагинацией
func apiPagedGet[T any](
	ctx context.Context,
	cli *http.Client,
	token string,
	urlBuilder func(pageToken string) string,
	field string,
	out *[]T,
) error {
	pageToken := ""
	for {
		url := urlBuilder(pageToken)

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

		var raw map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return fmt.Errorf("decode failed: %w", err)
		}

		items, ok := raw[field].([]any)
		if !ok {
			break
		}

		b, _ := json.Marshal(items)
		var chunk []T
		if err := json.Unmarshal(b, &chunk); err != nil {
			return err
		}

		*out = append(*out, chunk...)

		next, _ := raw["nextPageToken"].(string)
		if next == "" {
			break
		}
		pageToken = next
	}
	return nil
}
