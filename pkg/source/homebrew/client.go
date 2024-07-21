package homebrew

import (
	"encoding/json"
	"fmt"
	"github.com/ekristen/distillery/pkg/common"
	"github.com/sirupsen/logrus"
	"net/http"
)

func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}

	return &Client{
		client: client,
	}
}

type Client struct {
	client *http.Client
}

func (h *Client) GetFormula(formula string) (*Formula, error) {
	url := fmt.Sprintf("https://formulae.brew.sh/api/formula/%s.json", formula)

	logrus.Debugf("fetching formula: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", common.NAME, common.AppVersion))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var f *Formula
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return nil, err
	}

	return f, nil
}
