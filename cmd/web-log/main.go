package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"web-log/internal/history"
	"web-log/internal/summary"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		runTags(os.Args[1:])
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "tags":
		runTags(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Println(version)
	case "help", "--help", "-h":
		printHelp()
	default:
		if strings.HasPrefix(cmd, "-") {
			runTags(os.Args[1:])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func runTags(args []string) {
	fs := flag.NewFlagSet("tags", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to summarize")
	from := fs.String("from", "", "Start date (YYYY-MM-DD)")
	to := fs.String("to", "", "End date (YYYY-MM-DD)")
	dedupe := fs.Bool("dedupe", true, "Deduplicate URLs")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	since, until, startDate, endDate, actualDays, err := summary.DateRange(*days, *from, *to)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	entries, errs := history.ReadAllHistory(&since, &until)
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}

	if *dedupe {
		entries = history.Deduplicate(entries)
	}

	if len(entries) == 0 {
		fmt.Println("No browsing history found for this period.")
		return
	}

	output, err := summary.TagsSummary(entries, startDate, endDate, actualDays)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(output)
}

func printHelp() {
	fmt.Println("web-log — browsing history summary")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  web-log tags [--days N] [--from YYYY-MM-DD] [--to YYYY-MM-DD]")
	fmt.Println("  web-log (same as tags)")
	fmt.Println("  web-log version")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  web-log")
	fmt.Println("  web-log tags --days 7")
	fmt.Println("  web-log tags --from 2026-01-01 --to 2026-01-31")
}
