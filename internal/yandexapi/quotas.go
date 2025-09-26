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

// --- список сервисов ---
type quotaService struct {
	ID string `json:"id"`
}

func (c *client) ListQuotaServices() ([]string, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var services []quotaService
	err = apiPagedGet(
		context.Background(),
		c.httpCli,
		token,
		func(pageToken string) string {
			if pageToken == "" {
				return "https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits/services?resourceType=resource-manager.cloud&pageSize=1000"
			}
			return fmt.Sprintf(
				"https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits/services?resourceType=resource-manager.cloud&pageSize=1000&pageToken=%s",
				pageToken,
			)
		},
		"services",
		&services,
	)
	if err != nil {
		return nil, fmt.Errorf("list quota services failed: %w", err)
	}

	out := make([]string, 0, len(services))
	for _, s := range services {
		if s.ID != "" {
			out = append(out, s.ID)
		}
	}
	return out, nil
}

// --- список квот ---
func (c *client) ListQuotaLimits(cloudID, service string) ([]Quota, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var quotas []Quota
	err = apiPagedGet(
		context.Background(),
		c.httpCli,
		token,
		func(pageToken string) string {
			if pageToken == "" {
				return fmt.Sprintf(
					"https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits?resource.id=%s&resource.type=resource-manager.cloud&service=%s&pageSize=1000",
					cloudID, service,
				)
			}
			return fmt.Sprintf(
				"https://quota-manager.api.cloud.yandex.net/quota-manager/v1/quotaLimits?resource.id=%s&resource.type=resource-manager.cloud&service=%s&pageSize=1000&pageToken=%s",
				cloudID, service, pageToken,
			)
		},
		"quotaLimits",
		&quotas,
	)
	if err != nil {
		return nil, fmt.Errorf("list quota limits failed: %w", err)
	}

	return quotas, nil
}
