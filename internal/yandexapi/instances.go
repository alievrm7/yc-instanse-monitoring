package yandexapi

import (
	"context"
	"fmt"
)

type Instance struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	CloudID    string
	IPInternal string
	IPExternal string
}

type listInstancesResp struct {
	Instances []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
		NetIfs []struct {
			PrimaryV4Address struct {
				Address     string `json:"address"`
				OneToOneNat struct {
					Address string `json:"address"`
				} `json:"oneToOneNat"`
			} `json:"primaryV4Address"`
		} `json:"networkInterfaces"`
	} `json:"instances"`
}

func (c *client) ListInstancesByCloud(cloudID string) ([]Instance, error) {
	folders, err := c.ListFolders(cloudID)
	if err != nil {
		return nil, fmt.Errorf("list folders failed: %w", err)
	}

	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var all []Instance
	for _, f := range folders {
		var resp listInstancesResp
		url := fmt.Sprintf("https://compute.api.cloud.yandex.net/compute/v1/instances?folderId=%s", f.ID)

		if err := apiGet(context.Background(), c.httpCli, token, url, &resp); err != nil {
			return nil, err
		}

		for _, i := range resp.Instances {
			ipInternal, ipExternal := "", ""
			if len(i.NetIfs) > 0 {
				ipInternal = i.NetIfs[0].PrimaryV4Address.Address
				if i.NetIfs[0].PrimaryV4Address.OneToOneNat.Address != "" {
					ipExternal = i.NetIfs[0].PrimaryV4Address.OneToOneNat.Address
				}
			}
			all = append(all, Instance{
				ID:         i.ID,
				Name:       i.Name,
				Status:     i.Status,
				CloudID:    cloudID,
				IPInternal: ipInternal,
				IPExternal: ipExternal,
			})
		}
	}
	return all, nil
}
