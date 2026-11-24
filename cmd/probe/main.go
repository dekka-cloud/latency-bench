// Command probe measures latency-bench targets from where it runs and writes the result as
// JSON for report to pool with others.
//
//	probe --targets targets.json --city Accra --network MTN --access 4g
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/dekka-cloud/latency-bench/internal/probe"
	"github.com/dekka-cloud/latency-bench/internal/results"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "probe:", err)
		os.Exit(1)
	}
}

func run() error {
	o := probe.Defaults()
	o.Version = version
	o.Progress = os.Stderr
	targetsFile := flag.String("targets", "targets.json", `a JSON list of targets: [{"name": "...", "url": "https://..."}]`)
	out := flag.String("out", "", "where to write the result; results-<time>.json when unset")
	geoip := flag.String("geoip", "", "an MMDB country database, such as DB-IP's free IP to Country Lite, to name the countries a route crosses")
	flag.StringVar(&o.Vantage.City, "city", "", "the city you're measuring from, such as Accra")
	flag.StringVar(&o.Vantage.Country, "country", "", "its two-letter country code, such as GH")
	flag.StringVar(&o.Vantage.Network, "network", "", "your network, such as MTN, Telecel, AT or your broadband provider")
	flag.StringVar(&o.Vantage.Access, "access", "", "fibre, cable, dsl, 5g, 4g, 3g, satellite or other")
	flag.StringVar(&o.Vantage.Note, "note", "", "anything else worth recording, such as evening peak")
	flag.IntVar(&o.Rounds, "rounds", o.Rounds, "fresh connections and page loads per target")
	flag.BoolVar(&o.Insecure, "insecure", false, "accept self-signed certificates, as targets without a domain use")
	noUDP := flag.Bool("no-udp", false, "skip the UDP jitter and loss test")
	noRoute := flag.Bool("no-route", false, "skip tracing the route")
	flag.Parse()
	o.UDP, o.Route = !*noUDP, !*noRoute

	if o.Vantage.City == "" || o.Vantage.Network == "" || o.Vantage.Access == "" {
		return errors.New("say where you're measuring from with --city, --network and --access, so results can be compared")
	}
	targets, err := readTargets(*targetsFile)
	if err != nil {
		return err
	}
	if *geoip != "" {
		db, err := maxminddb.Open(*geoip)
		if err != nil {
			return fmt.Errorf("couldn't open the GeoIP database: %w", err)
		}
		defer func() { _ = db.Close() }()
		o.GeoIP = db
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	result := probe.Run(ctx, targets, o)

	path := *out
	if path == "" {
		path = fmt.Sprintf("results-%s.json", result.Started.Format("20060102-150405"))
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil { //nolint:gosec // a result file the person asked for
		return err
	}
	summary(result)
	fmt.Fprintf(os.Stderr, "\nWrote %s. Run report on several files to compare networks and days.\n", path)
	return nil
}

func readTargets(path string) ([]probe.Target, error) {
	data, err := os.ReadFile(path) //nolint:gosec // the person names this file
	if err != nil {
		return nil, fmt.Errorf("couldn't read the targets file: %w", err)
	}
	var targets []probe.Target
	if err := json.Unmarshal(data, &targets); err != nil {
		return nil, fmt.Errorf("%s isn't a JSON list of targets: %w", path, err)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("%s lists no targets", path)
	}
	for i := range targets {
		targets[i].URL = strings.TrimRight(targets[i].URL, "/")
	}
	return targets, nil
}

func summary(r results.Run) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "\nTARGET\tFIRST CONNECTION\tROUND TRIP\t20 CALLS\tPAGE + 20 CALLS\t1 MB DOWN\tUDP LOSS")
	for _, t := range r.Targets {
		if t.Error != "" {
			fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Error)
			continue
		}
		var connect, page []float64
		for _, c := range t.Connect {
			if c.Error == "" {
				connect = append(connect, c.Total)
			}
		}
		for _, p := range t.Page {
			if p.Calls == 20 {
				page = append(page, p.Total)
			}
		}
		calls := 0.0
		for _, c := range t.Chatty {
			if c.Calls == 20 {
				calls = c.Total
			}
		}
		down := ""
		for _, d := range t.Download {
			if d.Bytes == 1<<20 && d.Error == "" {
				down = fmt.Sprintf("%.1f Mbps", d.Mbps)
			}
		}
		loss := "n/a"
		if t.UDP != nil {
			loss = fmt.Sprintf("%.0f%%", t.UDP.Loss*100)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", t.Name, ms(results.Summarize(connect).Median), ms(t.Warm.Median), ms(calls),
			ms(results.Summarize(page).Median), down, loss)
	}
	_ = w.Flush()
}

func ms(v float64) string {
	switch {
	case v == 0 || math.IsNaN(v):
		return "n/a"
	case v < 10:
		return fmt.Sprintf("%.1f ms", v)
	default:
		return fmt.Sprintf("%.0f ms", v)
	}
}
