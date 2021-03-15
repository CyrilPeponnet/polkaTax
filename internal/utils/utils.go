package utils

import (
	"fmt"
	"time"

	"github.com/CyrilPeponnet/polkaTax/internal/coinmarket"
)

// FindClosestQuoteIndex O(log n) quote index lookup
func FindClosestQuoteIndex(t time.Time, data coinmarket.HistoricalData) (int, error) {

	var lookup func(lowerBound, upperBound int) (int, error)

	lookup = func(lowerBound, upperBound int) (int, error) {

		if upperBound == 0 {
			return upperBound, nil
		}

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
			}
			return lookup(mid, upperBound)

		}
		return -1, fmt.Errorf("Unable to find a matching timestamps")

	}

	return lookup(0, len(data)-1)
}

// abs return the abolute time duration
func abs(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
