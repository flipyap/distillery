package hashicorp

import (
	"encoding/json"
	"fmt"
	"github.com/ekristen/distillery/pkg/common"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	if client == nil {
		client = &http.Client{}
	}

	return &Client{
		client: client,
	}
}

// ListProducts lists all products
func (c *Client) ListProducts() (Products, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.releases.hashicorp.com/v1/products", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", common.NAME, common.AppVersion))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data Products
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) ListReleases(product string) ([]*Release, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.releases.hashicorp.com/v1/releases/%s?license_class=oss", product), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", common.NAME, common.AppVersion))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []*Release
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) GetVersion(product, version string) (*Release, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.releases.hashicorp.com/v1/releases/%s/%s?license_class=oss", product, version), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", common.NAME, common.AppVersion))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data *Release
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
