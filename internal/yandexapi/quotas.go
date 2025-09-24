package yandexapi

import (
	"context"
	"fmt"
)

type Quota struct {
	QuotaID string   `json:"quotaId"`
	Limit   *float64 `json:"limit"`
	Usage   *float64 `json:"usage"`
}

type quotaServicesResp struct {
	Services []struct {
		ID string `json:"id"`
	} `json:"services"`
}

func (c *client) ListQuotaServices() ([]string, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var resp quotaServicesResp
	url := "https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits/services?resourceType=resource-manager.cloud"

	if err := apiGet(context.Background(), c.httpCli, token, url, &resp); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(resp.Services))
	for _, s := range resp.Services {
		if s.ID != "" {
			out = append(out, s.ID)
		}
	}
	return out, nil
}

type quotaLimitsResp struct {
	QuotaLimits []Quota `json:"quotaLimits"`
}

func (c *client) ListQuotaLimits(cloudID, service string) ([]Quota, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var resp quotaLimitsResp
	url := fmt.Sprintf(
		"https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits?resource.id=%s&resource.type=resource-manager.cloud&service=%s",
		cloudID, service,
	)

	if err := apiGet(context.Background(), c.httpCli, token, url, &resp); err != nil {
		return nil, err
	}
	return resp.QuotaLimits, nil
}
