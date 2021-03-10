package main

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/CyrilPeponnet/polkaTax/internal/coinmarket"
	"github.com/CyrilPeponnet/polkaTax/internal/configuration"
	"github.com/CyrilPeponnet/polkaTax/internal/polkadot"
	"github.com/CyrilPeponnet/polkaTax/internal/print"
)

func main() {
	cfg := configuration.NewConfiguration()

	var err error

	var rewards polkadot.Rewards

	rewardchan := make(chan *polkadot.Reward, 500)
	rewardDone := make(chan bool)

	// Pump rewards to reward channel as they come (100 by 100)
	fmt.Printf("Getting all rewards for %s on %s network with era %s...", cfg.Account, cfg.Network, cfg.Era)

	go func() {
		defer duration(track(fmt.Sprintf("took")))

		err = polkadot.RewardsList(cfg, rewardchan)
		if err != nil {
			log.Fatalf("unable to get rewards: %s", err)
		}
		rewardDone <- true
	}()

	// Drain the rewards into concurrent jobs to get block time
	var wg sync.WaitGroup
	wg.Add(cfg.Concurency)

	over := false
	for i := 0; i < cfg.Concurency; i++ {
		go func() {
			defer wg.Done()
			for {
				select {

				case rw := <-rewardchan:
					ts, err := polkadot.BlockTime(cfg.BaseURL, cfg.Network, rw.BlockID)
					if err != nil {
						log.Fatalf("unable to get block timestamp: %s", err)
					}

					if !cfg.Start.IsZero() {
						if cfg.Start.After(ts) {
							continue
						}
					}

					if !cfg.End.IsZero() {
						if cfg.End.Before(ts) {
							continue
						}
					}

					rw.RewardTimeStamp = ts
					rewards = append(rewards, rw)

				case <-rewardDone:
					over = true
				default:

					if over && len(rewards) > 0 {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()

	// Sort then by date
	sort.Sort(rewards)

	// Get historical data

	//TODO: the limit is 10k points per queries, so we try to "group" our rewards by same interval to craft a
	// set of query to make ex: one reward every 6 hours for 5 consecutive days == 1 query

	data, err := coinmarket.GetHistoricalData(cfg.Network, rewards[0].RewardTimeStamp.Add(-24*time.Hour), rewards[len(rewards)-1].RewardTimeStamp.Add(24*time.Hour), cfg.Era)
	if err != nil {
		log.Fatalf("unable to get historical data: %s", err)
	}

	// For each reward try to get the closest quote
	for ridx := range rewards {

		i, err := findClosestQuoteIndex(rewards[ridx].RewardTimeStamp, data)
		if err != nil {
			log.Fatalf("unable to get quote for %s: %s", rewards[ridx].RewardTimeStamp, err)
		}
		rewards[ridx].USDQuoteTimeStamp = data[i].TimeStamp
		rewards[ridx].USDQuote = data[i].Value
	}

	// output
	print.FormattedOutput(rewards, cfg.CSV)

}

// findClosestQuoteIndex O(log n) quote index lookup
func findClosestQuoteIndex(t time.Time, data coinmarket.HistoricalData) (int, error) {

	var lookup func(lowerBound, upperBound int) (int, error)

	lookup = func(lowerBound, upperBound int) (int, error) {

		if upperBound >= lowerBound {

			mid := lowerBound + ((upperBound - lowerBound) / 2)

			// Return the closest item once we narrow it down to two items
			if upperBound-lowerBound == 1 {
				if abs(t.Sub(data[lowerBound].TimeStamp)) < abs(t.Sub(data[upperBound].TimeStamp)) {
					return lowerBound, nil
				}
				return upperBound, nil
			}

			if data[mid].TimeStamp.After(t) {
				return lookup(lowerBound, mid)
			} else {
				return lookup(mid, upperBound)
			}

		} else {
			return -1, fmt.Errorf("Unable to find a matching timestamps")
		}

	}

	return lookup(0, len(data)-1)
}

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func duration(msg string, start time.Time) {
	fmt.Printf("%v: %v\n", msg, time.Since(start))
}

func abs(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
