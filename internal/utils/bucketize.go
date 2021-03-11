package utils

import (
	"time"

	"github.com/CyrilPeponnet/polkaTax/internal/polkadot"
)

// Bucketize will try to bucketize the consecutize rewards given the era
// interface to minimize the number of requests made to get quotes
func Bucketize(data polkadot.Rewards, era time.Duration) []polkadot.Rewards {

	buckets := []polkadot.Rewards{}

	bucket := polkadot.Rewards{}

	for idx := range data {

		// if the timestamp is invalid we drop it.
		if data[idx].RewardTimeStamp.IsZero() {
			continue
		}

		// add item to bucket
		bucket = append(bucket, data[idx])

		// Is the next item within era
		if idx < len(data)-2 {

			// If the next item is not within the era we create a new bucket
			if abs(data[idx].RewardTimeStamp.Sub(data[idx+1].RewardTimeStamp)) > era+era/10 {
				buckets = append(buckets, bucket)
				bucket = polkadot.Rewards{}
			}
		}

	}

	buckets = append(buckets, bucket)

	return buckets

}
