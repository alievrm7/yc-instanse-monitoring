package yandexapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Cloud struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listCloudsResp struct {
	Clouds []Cloud `json:"clouds"`
}

func (c *client) ListClouds() ([]Cloud, error) {
	url := "https://resource-manager.api.cloud.yandex.net/resource-manager/v1/clouds"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data listCloudsResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	return data.Clouds, nil
}
