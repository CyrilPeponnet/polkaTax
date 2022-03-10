package main

import (
	"fmt"
	"os"
	"sort"
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

	rewards, err = polkadot.RewardsList(cfg)
	if err != nil {
		fmt.Println(goterm.Color(fmt.Sprintf("unable to get rewards: %s", err), goterm.RED))
		os.Exit(1)
	}

	sort.Sort(rewards)

	return rewards
}

func quoteRewards(unquotedRewards polkadot.Rewards, cfg *configuration.Configuration) (quotedRewards polkadot.Rewards) {

	// Bucketize rewards per contiguous era
	// Get USD quote for each bucket
	start := time.Now()

	// Get historical data, if they are not complete we go back to adjust the time window
	var data coinmarket.HistoricalData

	// set the boundaries with +-24h
	requestedLowerBound := unquotedRewards[0].RewardTimeStamp.Add(-24 * time.Hour)
	requestedUpperbound := unquotedRewards[len(unquotedRewards)-1].RewardTimeStamp.Add(24 * time.Hour)

	lowerBound := requestedLowerBound
	upperbound := requestedUpperbound

	for {

		fmt.Printf("\rGetting historical data from %s to %s...\n", lowerBound, upperbound)

		d, err := coinmarket.GetHistoricalData(cfg.Network, lowerBound, upperbound)
		if err != nil {
			fmt.Println(goterm.Color(fmt.Sprintf("unable to get historical data: %s", err), goterm.RED))
			os.Exit(1)
		}

		if len(d) == 0 {
			break
		}

		data = append(data, d...)

		if d[0].TimeStamp.Before(requestedLowerBound) {
			break
		}

		if d[0].TimeStamp.After(requestedLowerBound) {
			// upperbound is now TS for first data we get
			upperbound = d[0].TimeStamp.Add(-12 * time.Hour)
			continue
		}

		break

	}

	sort.Sort(data)

	// For each reward try to get the closest quote
	for ridx := range unquotedRewards {

		i, err := utils.FindClosestQuoteIndex(unquotedRewards[ridx].RewardTimeStamp, data)
		if err != nil {
			fmt.Println(goterm.Color(fmt.Sprintf("unable to get quote for %s: %s", unquotedRewards[ridx].RewardTimeStamp, err), goterm.RED))
			os.Exit(1)
		}

		unquotedRewards[ridx].USDQuoteTimeStamp = data[i].TimeStamp
		unquotedRewards[ridx].USDQuote = data[i].Value
		quotedRewards = append(quotedRewards, unquotedRewards[ridx])
	}

	fmt.Println(goterm.Color(fmt.Sprintf("it took %s to process %d rewards", time.Since(start), len(quotedRewards)), goterm.GREEN))

	return quotedRewards
}
