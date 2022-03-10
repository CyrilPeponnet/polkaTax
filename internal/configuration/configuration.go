// Package configuration is a small package for handling configuration
package configuration

import (
	"fmt"
	"os"
	"time"

	"github.com/buger/goterm"
	"go.aporeto.io/addedeffect/lombric"
)

// Configuration hold the service configuration.
type Configuration struct {
	Account    string `mapstructure:"account" desc:"The account address to use" required:"true"`
	Network    string `mapstructure:"network" desc:"The network to use" default:"Polkadot" allowed:"Polkadot,Kusama"`
	BaseURL    string `mapstructure:"url" desc:"The scanner api url to use" default:"https://api.dotscanner.com"`
	From       string `mapstructure:"from" desc:"Optional starting date"`
	To         string `mapstructure:"to" desc:"Optional ending date"`
	CSV        string `mapstructure:"csv" desc:"To save the results to a csv file"`
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
			fmt.Println(goterm.Color(fmt.Sprintf("unable to parse %s as date", c.From), goterm.RED))
			os.Exit(1)
		}
	}

	if c.To != "" {
		if c.End, err = time.Parse(time.RFC3339, c.To); err != nil {
			fmt.Println(goterm.Color(fmt.Sprintf("unable to parse %s as date", c.To), goterm.RED))
			os.Exit(1)
		}
	} else if !c.Start.IsZero() {
		c.End = time.Now()
	}

	if !c.Start.IsZero() && !c.End.IsZero() {
		if c.Start.After(c.End) {
			fmt.Println(goterm.Color(fmt.Sprintf("from cannot be greater than to: %s > %s", c.Start, c.End), goterm.RED))
			os.Exit(1)
		}
	}

	if c.Start.IsZero() {
		fmt.Printf("Computing data for %s on %s network...\n", goterm.Bold(c.Account), goterm.Bold(c.Network))
	} else {
		fmt.Printf("Computing data for %s on %s network between %s and %s...\n", goterm.Bold(c.Account), goterm.Bold(c.Network), goterm.Bold(c.Start.Format(time.RFC3339)), goterm.Bold(c.End.Format(time.RFC3339)))
	}

	return c
}
