package coinmarket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const (
	dotCode       = "6636"
	ksmCode       = "5034"
	coinmarketURL = "https://web-api.coinmarketcap.com/v1.1/cryptocurrency/quotes/historical"
)

// coinMarketHistoricalData is the data structure returned
type coinMarketHistoricalData struct {
	Status struct {
		Timestamp    time.Time   `json:"timestamp"`
		ErrorCode    int         `json:"error_code"`
		ErrorMessage string      `json:"error_message"`
		Elapsed      int         `json:"elapsed"`
		CreditCount  int         `json:"credit_count"`
		Notice       interface{} `json:"notice"`
	} `json:"status"`
	MapQuotes map[time.Time]struct {
		USD []float64 `json:"USD"`
	} `json:"data"`
}

// USDQuote represents the USD quotation for a given time
type USDQuote struct {
	TimeStamp time.Time
	Value     float64
}

// HistoricalData is the list of USD quotes
type HistoricalData []USDQuote

func (p HistoricalData) Len() int {
	return len(p)
}

func (p HistoricalData) Less(i, j int) bool {
	return p[i].TimeStamp.Before(p[j].TimeStamp)
}

func (p HistoricalData) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// GetHistoricalData will retrieve historical data for a given currency and time window
func GetHistoricalData(network string, from time.Time, to time.Time, interval time.Duration) (HistoricalData, error) {

	chd := &coinMarketHistoricalData{}
	hd := HistoricalData{}

	// Set proper code
	code := dotCode
	if network == "kusama" {
		code = ksmCode
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	url, err := url.Parse(coinmarketURL)
	if err != nil {
		panic(err)
	}
	q := url.Query()
	q.Add("convert", "USD,BTC")
	q.Add("format", "chart_crypto_details")
	q.Add("id", code)
	q.Add("interval", fmt.Sprintf("%dh", int(interval.Hours())))
	q.Add("time_start", fmt.Sprintf("%d", from.Unix()))
	q.Add("time_end", fmt.Sprintf("%d", to.Unix()))

	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return hd, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return hd, fmt.Errorf("unable to send request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close() //nolint
	if err != nil {
		return hd, fmt.Errorf("unable to read body: %w", err)
	}

	err = json.Unmarshal(body, &chd)

	if err != nil {
		return hd, fmt.Errorf("unable to unmarshal data %s", err)
	}

	if len(chd.MapQuotes) == 0 {
		// workaround:
		if chd.Status.ErrorMessage == `"time_start" must be older than "time_end".` {
			return hd, fmt.Errorf("unable to get historical data from %s because of error %s", url.String(), fmt.Sprintf("unable to find quote in the %s to %s range", from, to))
		}
		return hd, fmt.Errorf("unable to get historical data from %s because of error %s", url.String(), chd.Status.ErrorMessage)
	}

	// convert to HistoricalData
	for ts, values := range chd.MapQuotes {
		hd = append(hd, USDQuote{TimeStamp: ts.UTC(), Value: values.USD[0]})
	}

	sort.Sort(hd)

	return hd, nil
}
