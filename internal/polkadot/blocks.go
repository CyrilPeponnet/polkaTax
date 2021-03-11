package polkadot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type block struct {
	Errors []interface{} `json:"errors"`
	Data   struct {
		ID         int `json:"id"`
		Attributes struct {
			Datetime time.Time `json:"datetime"`
		} `json:"attributes"`
	} `json:"data"`
}

// BlockTime will retrieve a block time
func BlockTime(burl string, network string, id int) (time.Time, error) {

	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	// Build the url
	url, err := url.Parse(burl)
	if err != nil {
		return time.Now(), err
	}
	url.Path = fmt.Sprintf("api/v1/%s/block/%d", network, id)

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return time.Now(), fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return time.Now(), fmt.Errorf("unable to send request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close() //nolint
	if err != nil {
		return time.Now(), fmt.Errorf("unable to read body: %w", err)
	}

	b := block{}
	err = json.Unmarshal(body, &b)

	if err != nil {
		return time.Now(), fmt.Errorf("uanble to unmarshal data %s", err)
	}

	return b.Data.Attributes.Datetime.UTC(), nil
}
