package yandexapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Folder struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CloudID string `json:"cloudId"`
}

// ответ API
type listFoldersResp struct {
	Folders []Folder `json:"folders"`
}

func (c *client) ListFolders(cloudID string) ([]Folder, error) {
	url := fmt.Sprintf("https://resource-manager.api.cloud.yandex.net/resource-manager/v1/folders?cloudId=%s", cloudID)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list folders request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data listFoldersResp
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode folders failed: %w", err)
	}

	return data.Folders, nil
}
