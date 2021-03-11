package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CyrilPeponnet/polkaTax/internal/coinmarket"
	"github.com/CyrilPeponnet/polkaTax/internal/configuration"
	"github.com/CyrilPeponnet/polkaTax/internal/polkadot"
	"github.com/CyrilPeponnet/polkaTax/internal/utils"
	"github.com/buger/goterm"
)

func main() {
	cfg := configuration.NewConfiguration()

	// output
	utils.FormattedOutput(quoteRewards(retrieveRewards(cfg), cfg), cfg.CSV)

}

func retrieveRewards(cfg *configuration.Configuration) (rewards polkadot.Rewards) {

	var err error

	start := time.Now()

	rewards, err = polkadot.RewardsList(cfg)
	if err != nil {
		fmt.Println(goterm.Color(fmt.Sprintf("unable to get rewards: %s", err), goterm.RED))
		os.Exit(1)
	}

	jobs := make(chan int)

	var wg sync.WaitGroup

	total := len(rewards)
	// Start our workers
	var ops uint64
	for i := 0; i < cfg.Concurency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rw := range jobs {

				atomic.AddUint64(&ops, 1)

				fmt.Printf("\r* Exploring block %d (%d/%d)", rewards[rw].BlockID, atomic.LoadUint64(&ops), total)

				ts, err := polkadot.BlockTime(cfg.BaseURL, cfg.Network, rewards[rw].BlockID)
				if err != nil {
					fmt.Println(goterm.Color(fmt.Sprintf("unable to get block timestamp: %s", err), goterm.RED))
					os.Exit(1)
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
				rewards[rw].RewardTimeStamp = ts

			}

		}()
	}

	// Pump the rewards to the job channel
	for ridx := range rewards {
		jobs <- ridx
	}

	close(jobs)
	wg.Wait()

	sort.Sort(rewards)
	fmt.Println(goterm.Color(fmt.Sprintf("\n took %s", time.Since(start)), goterm.GREEN))
	return rewards
}

func quoteRewards(unquotedRewards polkadot.Rewards, cfg *configuration.Configuration) (quotedRewards polkadot.Rewards) {

	// Bucketize rewards per contiguous era
	// Get USD quote for each bucket
	buckets := utils.Bucketize(unquotedRewards, cfg.Era)
	start := time.Now()

	for i, bucket := range buckets {

		fmt.Printf("\r* Getting historical data for bucket %s/%s...", goterm.Bold(fmt.Sprintf("%d", i)), goterm.Bold(fmt.Sprintf("%d", len(buckets)-1)))

		data, err := coinmarket.GetHistoricalData(cfg.Network, bucket[0].RewardTimeStamp.Add(-24*time.Hour), bucket[len(bucket)-1].RewardTimeStamp.Add(24*time.Hour), cfg.Era)
		if err != nil {
			fmt.Println(goterm.Color(fmt.Sprintf("unable to get historical data: %s", err), goterm.RED))
			os.Exit(1)
		}

		// For each reward try to get the closest quote
		for ridx := range bucket {

			i, err := utils.FindClosestQuoteIndex(bucket[ridx].RewardTimeStamp, data)
			if err != nil {
				fmt.Println(goterm.Color(fmt.Sprintf("unable to get quote for %s: %s", bucket[ridx].RewardTimeStamp, err), goterm.RED))
				os.Exit(1)
			}

			bucket[ridx].USDQuoteTimeStamp = data[i].TimeStamp
			bucket[ridx].USDQuote = data[i].Value
			quotedRewards = append(quotedRewards, bucket[ridx])
		}

	}

	fmt.Println(goterm.Color(fmt.Sprintf("took %s to process %d rewards", time.Since(start), len(quotedRewards)), goterm.GREEN))

	return quotedRewards
}
