package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/CyrilPeponnet/polkaTax/internal/polkadot"
	"github.com/buger/goterm"
	"github.com/olekukonko/tablewriter"
)

// FormattedOutput will format the output
func FormattedOutput(rewards polkadot.Rewards, csv string) {
	if csv != "" {
		outputcsv(rewards, csv)
	} else {
		outputTable(rewards)
	}
}

// FormattedOutput will format the output
func FormattedOutputForNotFoundRewards(notFoundRewards polkadot.Rewards) {
	outputTableForNotFoundRewards(notFoundRewards)
}

func outputcsv(rewards polkadot.Rewards, csv string) {

	f, err := os.Create(csv)
	if err != nil {
		fmt.Println(goterm.Color(fmt.Sprintf("unable to create file %s: %s", csv, err), goterm.RED))
		os.Exit(1)
	}

	defer f.Close() //nolint

	w := bufio.NewWriter(f)

	_, err = fmt.Fprintln(w, "REWARD DATE,AMOUNT,USD QUOTE DATE,USD QUOTE,USD VALUE")
	if err != nil {
		fmt.Println(goterm.Color(fmt.Sprintf("unable to write to %s: %s", csv, err), goterm.RED))
		os.Exit(1)
	}

	for _, r := range rewards {
		_, err = fmt.Fprintf(w, "%s,%f,%s,%f,%f\n", r.RewardTimeStamp.Format(time.RFC3339), r.Value, r.USDQuoteTimeStamp.Format(time.RFC3339), r.USDQuote, r.Value*r.USDQuote)
		if err != nil {
			fmt.Println(goterm.Color(fmt.Sprintf("unable to write to %s: %s", csv, err), goterm.RED))
			os.Exit(1)
		}
	}

	w.Flush() //nolint
}

func outputTable(rewards polkadot.Rewards) {

	header := []string{"Reward Date", "Amount", "USD Quote Date", "USD Quote", "USD Value"}

	rows := [][]string{}
	total := 0.0
	totalDOT := 0.0
	for _, r := range rewards {
		rows = append(rows, []string{r.RewardTimeStamp.Format(time.RFC3339), fmt.Sprintf("%f", r.Value), r.USDQuoteTimeStamp.Format(time.RFC3339), fmt.Sprintf("%f", r.USDQuote), fmt.Sprintf("%f", r.Value*r.USDQuote)})
		total = total + r.Value*r.USDQuote
		totalDOT += r.Value
	}

	out := &bytes.Buffer{}

	colors := make([]tablewriter.Colors, len(header))
	for i := 0; i < len(header); i++ {
		colors[i] = tablewriter.Color(tablewriter.FgCyanColor, tablewriter.Bold)
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderLine(true)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderColor(colors...)

	table.Render()

	fmt.Println("\n" + out.String())
	fmt.Println(fmt.Sprintf("Total USD for period: %s$ for %f DOT", goterm.Bold(fmt.Sprintf("%f", total)), totalDOT))
}

func outputTableForNotFoundRewards(rewards polkadot.Rewards) {

	header := []string{"Reward Date", "Amount"}

	rows := [][]string{}
	total := 0.0
	for _, r := range rewards {
		rows = append(rows, []string{r.RewardTimeStamp.Format(time.RFC3339), fmt.Sprintf("%f", r.Value)})
		total = total + r.Value
	}

	out := &bytes.Buffer{}

	colors := make([]tablewriter.Colors, len(header))
	for i := 0; i < len(header); i++ {
		colors[i] = tablewriter.Color(tablewriter.FgCyanColor, tablewriter.Bold)
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderLine(true)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderColor(colors...)

	table.Render()

	fmt.Println("\n" + out.String())
	fmt.Println("Total of DOT not quoted: ", goterm.Bold(fmt.Sprintf("%f", total)))
}
