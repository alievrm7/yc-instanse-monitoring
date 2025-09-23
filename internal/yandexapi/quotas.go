package yandexapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Quota struct {
	QuotaID string   `json:"quotaId"`
	Limit   *float64 `json:"limit"`
	Usage   *float64 `json:"usage"`
}

// ---- список сервисов ----
type quotaServicesResp struct {
	Services []struct {
		ID string `json:"id"`
	} `json:"services"`
}

func (c *client) ListQuotaServices() ([]string, error) {
	url := "https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits/services?resourceType=resource-manager.cloud"

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quota services request failed: %w", err)
	}
	defer resp.Body.Close()

	var data quotaServicesResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode quota services failed: %w", err)
	}

	out := make([]string, 0, len(data.Services))
	for _, s := range data.Services {
		if s.ID != "" {
			out = append(out, s.ID)
		}
	}
	return out, nil
}

// ---- список квот ----
type quotaLimitsResp struct {
	QuotaLimits []Quota `json:"quotaLimits"`
}

func (c *client) ListQuotaLimits(cloudID, service string) ([]Quota, error) {
	url := fmt.Sprintf(
		"https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits?resource.id=%s&resource.type=resource-manager.cloud&service=%s",
		cloudID, service,
	)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("quota limits request failed: %w", err)
	}
	defer resp.Body.Close()

	var data quotaLimitsResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode quota limits failed: %w", err)
	}

	return data.QuotaLimits, nil
}
