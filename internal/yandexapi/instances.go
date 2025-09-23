package yandexapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		Zone   string `json:"zone_id"`
		Status string `json:"status"`
		FQDN   string `json:"fqdn"`
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

// Реализация метода интерфейса Client
func (c *client) ListInstancesByCloud(cloudID string) ([]Instance, error) {
	// Получаем список фолдеров для этого облака
	folders, err := c.ListFolders(cloudID)
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	if len(folders) == 0 {
		fmt.Printf("DEBUG: no folders found in cloud=%s\n", cloudID)
		return nil, nil
	}

	var all []Instance
	for _, f := range folders {

		url := fmt.Sprintf("https://compute.api.cloud.yandex.net/compute/v1/instances?folderId=%s", f.ID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		token, err := c.getToken()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.httpCli.Do(req)
		if err != nil {
			return nil, fmt.Errorf("list instances request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var data listInstancesResp
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("decode instances failed: %w", err)
		}

		for _, i := range data.Instances {
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
				CloudID:    cloudID, // 👈 подставляем ID
				IPInternal: ipInternal,
				IPExternal: ipExternal,
			})
		}
	}
	return all, nil
}
