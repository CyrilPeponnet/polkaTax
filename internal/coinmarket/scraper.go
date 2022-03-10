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
	coinmarketURL = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/historical"
)

type coinMarketHistoricalData struct {
	Data struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Quotes []struct {
			TimeOpen  time.Time `json:"timeOpen"`
			TimeClose time.Time `json:"timeClose"`
			TimeHigh  time.Time `json:"timeHigh"`
			TimeLow   time.Time `json:"timeLow"`
			Quote     struct {
				Open      float64   `json:"open"`
				High      float64   `json:"high"`
				Low       float64   `json:"low"`
				Close     float64   `json:"close"`
				Volume    float64   `json:"volume"`
				MarketCap float64   `json:"marketCap"`
				Timestamp time.Time `json:"timestamp"`
			} `json:"quote"`
		} `json:"quotes"`
	} `json:"data"`
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    string    `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      string    `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
	} `json:"status"`
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
func GetHistoricalData(network string, from time.Time, to time.Time) (HistoricalData, error) {

	chd := &coinMarketHistoricalData{}
	hd := HistoricalData{}

	// Set proper code
	code := dotCode
	if network == "Kusama" {
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
	q.Add("id", code)
	q.Add("convertId", "2781")
	q.Add("timeStart", fmt.Sprintf("%d", from.Unix()))
	q.Add("timeEnd", fmt.Sprintf("%d", to.Unix()))

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
		fmt.Println(url.String())
		fmt.Println(string(body))
		return hd, fmt.Errorf("unable to unmarshal data %s", err)
	}

	if len(chd.Data.Quotes) == 0 {
		// workaround:
		if chd.Status.ErrorMessage == `"time_start" must be older than "time_end".` {
			return hd, fmt.Errorf("unable to get historical data from %s because of error %s", url.String(), fmt.Sprintf("unable to find quote in the %s to %s range", from, to))
		}
		return hd, nil
	}

	// convert to HistoricalData
	for _, quote := range chd.Data.Quotes {
		hd = append(hd, USDQuote{TimeStamp: quote.TimeOpen.UTC(), Value: quote.Quote.Open})
		hd = append(hd, USDQuote{TimeStamp: quote.TimeClose.UTC(), Value: quote.Quote.Close})
		hd = append(hd, USDQuote{TimeStamp: quote.TimeHigh.UTC(), Value: quote.Quote.High})
		hd = append(hd, USDQuote{TimeStamp: quote.TimeLow.UTC(), Value: quote.Quote.Low})
	}

	sort.Sort(hd)

	return hd, nil
}
