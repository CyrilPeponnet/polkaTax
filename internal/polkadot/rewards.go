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

// Reward preseents the reward for a given time
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
	Data   struct {
		Rewards []struct {
			Amount    int64  `json:"amount"`
			TimeStamp string `json:"timestamp"`
		} `json:"rewards"`
	} `json:"data"`
}

// This is to convert back to dot
const dotDivider = 10000000000
const ksmDivider = 1000000000000

// RewardsList will retrieve events related to stacking
func RewardsList(cfg *configuration.Configuration) (Rewards, error) {

	rewards := Rewards{}

	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	// Starting page
	divider := ksmDivider
	if cfg.Network == "Polkadot" {
		divider = dotDivider
	}

	// Build the url
	url, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return rewards, err
	}
	url.Path = fmt.Sprintf("accounts/%v/rewards", cfg.Account)

	q := url.Query()
	q.Set("chain", cfg.Network)

	// Loop until we are done (no results) or if time windows is exhautsted

	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return rewards, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return rewards, fmt.Errorf("unable to send request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close() //nolint
	if err != nil {
		return rewards, fmt.Errorf("unable to read body: %w", err)
	}

	ev := Events{}
	err = json.Unmarshal(body, &ev)

	if err != nil {
		return rewards, fmt.Errorf("uanble to unmarshal data %s", err)
	}

	// no more events means we are done
	if len(ev.Data.Rewards) == 0 {
		return rewards, fmt.Errorf("No rewards founds for %s", cfg.Account)
	}

	// Retrieve the date time from the block
	for _, r := range ev.Data.Rewards {

		t, err := time.Parse("2-Jan-2006 5:04:05 PM-07:00", r.TimeStamp)
		if err != nil {
			return rewards, fmt.Errorf("error: unable to parse time:%w", err)
		}
		rewards = append(rewards, &Reward{Value: float64(r.Amount) / float64(divider), RewardTimeStamp: t})
	}

	return rewards, nil

}
