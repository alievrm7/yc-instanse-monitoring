package yandexapi

import (
	"context"
	"fmt"
)

type Folder struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CloudID string `json:"cloudId"`
}

type listFoldersResp struct {
	Folders []Folder `json:"folders"`
}

func (c *client) ListFolders(cloudID string) ([]Folder, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var resp listFoldersResp
	url := fmt.Sprintf("https://resource-manager.api.cloud.yandex.net/resource-manager/v1/folders?cloudId=%s", cloudID)

	if err := apiGet(context.Background(), c.httpCli, token, url, &resp); err != nil {
		return nil, err
	}
	return resp.Folders, nil
}
