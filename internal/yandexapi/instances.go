package yandexapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Instance — информация о виртуальной машине
type Instance struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Zone     string `json:"zone_id"`
	Status   string `json:"status"`
	CloudID  string
	Cloud    string
	FolderID string
	Folder   string
	Hostname string `json:"fqdn"`
	IP       string
}

// ---------------------------
// Запрос инстансов в папке
// ---------------------------
func (c *client) ListInstances(folderID string) ([]Instance, error) {
	url := fmt.Sprintf("https://compute.api.cloud.yandex.net/compute/v1/instances?folderId=%s", folderID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
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
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	out := []Instance{}
	for _, i := range data.Instances {
		ip := ""
		if len(i.NetIfs) > 0 {
			ip = i.NetIfs[0].PrimaryV4Address.Address
			if i.NetIfs[0].PrimaryV4Address.OneToOneNat.Address != "" {
				ip = i.NetIfs[0].PrimaryV4Address.OneToOneNat.Address
			}
		}
		out = append(out, Instance{
			ID:       i.ID,
			Name:     i.Name,
			Zone:     i.Zone,
			Status:   i.Status,
			Hostname: i.FQDN,
			IP:       ip,
		})
	}
	return out, nil
}

// ---------------------------
// Обход всех облаков/папок
// ---------------------------
func (c *client) ListAllInstances() ([]Instance, error) {
	clouds, err := c.ListClouds()
	if err != nil {
		return nil, err
	}

	var all []Instance
	for _, cloud := range clouds {
		folders, err := c.ListFolders(cloud.ID)
		if err != nil {
			return nil, err
		}

		for _, folder := range folders {
			instances, err := c.ListInstances(folder.ID)
			if err != nil {
				return nil, err
			}

			// проставляем Cloud/Folder name + ID
			for i := range instances {
				instances[i].CloudID = cloud.ID
				instances[i].Cloud = cloud.Name
				instances[i].FolderID = folder.ID
				instances[i].Folder = folder.Name
			}
			all = append(all, instances...)
		}
	}
	return all, nil
}
