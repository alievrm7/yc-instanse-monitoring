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

func (c *client) ListFolders(cloudID string) ([]Folder, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var all []Folder
	err = apiPagedGet(
		context.Background(),
		c.httpCli,
		token,
		func(pageToken string) string {
			if pageToken == "" {
				return fmt.Sprintf(
					"https://resource-manager.api.cloud.yandex.net/resource-manager/v1/folders?cloudId=%s&pageSize=1000",
					cloudID,
				)
			}
			return fmt.Sprintf(
				"https://resource-manager.api.cloud.yandex.net/resource-manager/v1/folders?cloudId=%s&pageSize=1000&pageToken=%s",
				cloudID, pageToken,
			)
		},
		"folders",
		&all,
	)
	if err != nil {
		return nil, fmt.Errorf("list folders failed: %w", err)
	}

	return all, nil
}
