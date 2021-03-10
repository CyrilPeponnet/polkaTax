// Package configuration is a small package for handling configuration
package configuration

import (
	"log"
	"time"

	"go.aporeto.io/addedeffect/lombric"
)

// Configuration hold the service configuration.
type Configuration struct {
	Account    string        `mapstructure:"account" desc:"The account address to use" required:"true"`
	Network    string        `mapstructure:"network" desc:"The network to use" default:"polkadot" allowed:"polkadot,kusama"`
	Era        time.Duration `mapstructure:"era" desc:"Optional set the era to compute historical quote. default is 6h for Kusama and 24h for Polkadot"`
	BaseURL    string        `mapstructure:"url" desc:"The KSM api url to use" default:"https://explorer-32.polkascan.io"`
	Concurency int           `mapstructure:"concurrent" desc:"The number of concurent jobs to run" default:"100"`
	From       string        `mapstructure:"from" desc:"Optional starting date"`
	To         string        `mapstructure:"to" desc:"Optional ending date"`
	CSV        string        `mapstructure:"csv" desc:"To save the results to a csv file"`
	Start, End time.Time
}

// Prefix returns the configuration prefix.
func (c *Configuration) Prefix() string { return "tracer" }

// NewConfiguration returns a new configuration.
func NewConfiguration() *Configuration {

	var err error

	c := &Configuration{}
	lombric.Initialize(c)

	if c.From != "" {
		if c.Start, err = time.Parse(time.RFC3339, c.From); err != nil {
			log.Fatalf("unable to parse %s as date", c.From)
		}
	}

	if c.To != "" {
		if c.End, err = time.Parse(time.RFC3339, c.To); err != nil {
			log.Fatalf("unable to parse %s as date", c.To)
		}
	}

	if !c.Start.IsZero() && !c.End.IsZero() {
		if c.Start.After(c.End) {
			log.Fatalf("from cannot be greater than to: %s > %s", c.Start, c.End)
		}
	}

	if c.Era == 0*time.Second {
		if c.Network == "polkadot" {
			c.Era = 24 * time.Hour
		} else {
			c.Era = 6 * time.Hour
		}
	}

	return c
}
