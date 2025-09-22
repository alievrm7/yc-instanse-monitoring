package yandexapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Folder struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CloudID string `json:"cloud_id"`
}

type listFoldersResp struct {
	Folders []Folder `json:"folders"`
}

func (c *client) ListFolders(cloudID string) ([]Folder, error) {
	url := fmt.Sprintf("https://resource-manager.api.cloud.yandex.net/resource-manager/v1/folders?cloudId=%s", cloudID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data listFoldersResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	return data.Folders, nil
}
