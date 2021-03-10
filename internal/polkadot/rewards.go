package polkadot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/CyrilPeponnet/polkaTax/internal/configuration"
)

// Rewards preseents the reward for a given time
type Reward struct {
	RewardTimeStamp, USDQuoteTimeStamp time.Time
	BlockID                            int
	Value                              float64
	USDQuote                           float64
}

// Rewards represent a sortable list of rewards
type Rewards []*Reward

func (p Rewards) Len() int {
	return len(p)
}

func (p Rewards) Less(i, j int) bool {
	return p[i].RewardTimeStamp.Before(p[j].RewardTimeStamp)
}

func (p Rewards) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Events represents the events reported by the events api
type Events struct {
	Errors []interface{} `json:"errors"`
	Data   []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			BlockID    int    `json:"block_id"`
			EvendID    string `json:"event_id"`
			Attributes []struct {
				Type      string      `json:"type"`
				Value     interface{} `json:"value"`
				OrigValue string      `json:"orig_value,omitempty"`
			} `json:"attributes"`
		} `json:"attributes"`
	} `json:"data"`
}

// search_index constants that represents the events
const search_index = "39"
const page_size = "200"
const evendID = "Reward"

// This is to convert back to dot
const dotDivider = 10000000000
const ksmDivider = 1000000000000

// RewardsList will retrieve events related to stacking
func RewardsList(cfg *configuration.Configuration, rewards chan<- *Reward) error {

	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	// Starting page
	page := 0
	divider := ksmDivider
	if cfg.Network == "polkadot" {
		divider = dotDivider
	}

	// Build the url
	url, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return err
	}
	url.Path = fmt.Sprintf("api/v1/%v/event", cfg.Network)

	q := url.Query()
	q.Set("filter[address]", cfg.Account)
	q.Set("filter[search_index]", search_index)
	q.Set("page[size]", page_size)

	// Loop until we are done (no results) or if time windows is exhautsted
	for {

		page++
		q.Set("page[number]", fmt.Sprintf("%d", page))
		url.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, url.String(), nil)
		if err != nil {
			return fmt.Errorf("unable to build request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("unable to send request: %w", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close() //nolint
		if err != nil {
			return fmt.Errorf("unable to read body: %w", err)
		}

		ev := Events{}
		err = json.Unmarshal(body, &ev)

		if err != nil {
			return fmt.Errorf("uanble to unmarshal data %s", err)
		}

		// no more events means we are done
		if len(ev.Data) == 0 {
			return nil
		}

		// Retrive the date time from the block
		for _, e := range ev.Data {
			// Keep only reward events
			if e.Attributes.EvendID != evendID {
				continue
			}
			var balance float64
			var ok bool
			for _, a := range e.Attributes.Attributes {
				if a.Type == "Balance" {
					if balance, ok = a.Value.(float64); !ok {
						return fmt.Errorf("error while converting balance to float64: %T", a.Value)
					}
				}
			}

			rewards <- &Reward{BlockID: e.Attributes.BlockID, Value: float64(balance / float64(divider))}
		}

	}

}
