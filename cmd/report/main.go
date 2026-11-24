// Command report pools probe results and prints a table per network: the median of each
// measurement against each target, across every run, as Markdown or CSV.
//
//	report results-*.json > report.md
//	report --csv results-*.json > report.csv
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/dekka-cloud/latency-bench/internal/results"
)

func main() {
	asCSV := flag.Bool("csv", false, "print CSV instead of Markdown")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "report: name one or more probe result files")
		os.Exit(2)
	}
	var runs []results.Run
	for _, path := range flag.Args() {
		data, err := os.ReadFile(path) //nolint:gosec // files the person names
		if err != nil {
			fail(err)
		}
		var r results.Run
		if err := json.Unmarshal(data, &r); err != nil {
			fail(fmt.Errorf("%s isn't a probe result: %w", path, err))
		}
		runs = append(runs, r)
	}
	rows := summarize(runs)
	if *asCSV {
		writeCSV(os.Stdout, rows)
		return
	}
	writeMarkdown(os.Stdout, rows, len(runs))
}

// row is one vantage and target, pooled across runs.
type row struct {
	vantage, target, location string
	runs                      int
	connect, tls, rtt, jitter float64
	calls20, page20           float64
	down1M, down10M, up1M     float64
	udpLoss                   float64
	hops                      float64
	countries                 string
}

type key struct{ vantage, target string }

type samples struct {
	location                          string
	runs                              int
	connect, tls, rtt, jitter         []float64
	calls20, page20                   []float64
	down1M, down10M, up1M, loss, hops []float64
	countries                         map[string]bool
}

func summarize(runs []results.Run) []row {
	pooled := map[key]*samples{}
	var order []key
	for _, r := range runs {
		v := strings.TrimSpace(fmt.Sprintf("%s, %s %s", r.Vantage.City, r.Vantage.Network, r.Vantage.Access))
		for _, t := range r.Targets {
			if t.Error != "" {
				continue
			}
			k := key{v, t.Name}
			s, ok := pooled[k]
			if !ok {
				s = &samples{location: t.Location, countries: map[string]bool{}}
				pooled[k] = s
				order = append(order, k)
			}
			s.runs++
			for _, c := range t.Connect {
				if c.Error == "" {
					s.connect = append(s.connect, c.Total)
					s.tls = append(s.tls, c.TLS)
				}
			}
			if t.Warm.N > 0 {
				s.rtt = append(s.rtt, t.Warm.Median)
				s.jitter = append(s.jitter, t.Warm.Jitter)
			}
			for _, c := range t.Chatty {
				if c.Calls == 20 {
					s.calls20 = append(s.calls20, c.Total)
				}
			}
			for _, p := range t.Page {
				if p.Calls == 20 {
					s.page20 = append(s.page20, p.Total)
				}
			}
			for _, d := range t.Download {
				if d.Error == "" && d.Bytes == 1<<20 {
					s.down1M = append(s.down1M, d.Mbps)
				}
				if d.Error == "" && d.Bytes == 10<<20 {
					s.down10M = append(s.down10M, d.Mbps)
				}
			}
			for _, u := range t.Upload {
				if u.Error == "" && u.Bytes == 1<<20 {
					s.up1M = append(s.up1M, u.Mbps)
				}
			}
			if t.UDP != nil {
				s.loss = append(s.loss, t.UDP.Loss)
			}
			if t.Route != nil {
				s.hops = append(s.hops, float64(len(t.Route.Hops)))
				for _, h := range t.Route.Hops {
					if h.Country != "" {
						s.countries[h.Country] = true
					}
				}
			}
		}
	}
	out := make([]row, 0, len(order))
	for _, k := range order {
		s := pooled[k]
		countries := make([]string, 0, len(s.countries))
		for c := range s.countries {
			countries = append(countries, c)
		}
		slices.Sort(countries)
		out = append(out, row{
			vantage: k.vantage, target: k.target, location: s.location, runs: s.runs,
			connect: median(s.connect), tls: median(s.tls), rtt: median(s.rtt), jitter: median(s.jitter),
			calls20: median(s.calls20), page20: median(s.page20),
			down1M: median(s.down1M), down10M: median(s.down10M), up1M: median(s.up1M),
			udpLoss: median(s.loss), hops: median(s.hops), countries: strings.Join(countries, " "),
		})
	}
	return out
}

// median is NaN when nothing was measured, which the tables print as n/a.
func median(v []float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	sorted := slices.Clone(v)
	slices.Sort(sorted)
	return results.Quantile(sorted, 0.5)
}

func writeMarkdown(w io.Writer, rows []row, runs int) {
	fmt.Fprintf(w, "Medians across %d probe runs. Times are in milliseconds and speeds in megabits per second.\n", runs)
	last := ""
	for _, r := range rows {
		if r.vantage != last {
			fmt.Fprintf(w, "\n### From %s\n\n", r.vantage)
			fmt.Fprintln(w, "| Target | Where | First connection | TLS | Round trip | Jitter | 20 calls | Page and 20 calls | 1 MB down | 10 MB down | 1 MB up | UDP loss | Hops | Countries crossed |")
			fmt.Fprintln(w, "| --- | --- | --: | --: | --: | --: | --: | --: | --: | --: | --: | --: | --: | --- |")
			last = r.vantage
		}
		fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			r.target, r.location, num(r.connect, 0), num(r.tls, 0), num(r.rtt, 1), num(r.jitter, 1), num(r.calls20, 0), num(r.page20, 0),
			num(r.down1M, 1), num(r.down10M, 1), num(r.up1M, 1), percent(r.udpLoss), num(r.hops, 0), r.countries)
	}
}

func writeCSV(w io.Writer, rows []row) {
	c := csv.NewWriter(w)
	_ = c.Write([]string{"vantage", "target", "location", "runs", "connect_ms", "tls_ms", "rtt_ms", "jitter_ms", "calls20_ms", "page20_ms",
		"down_1mb_mbps", "down_10mb_mbps", "up_1mb_mbps", "udp_loss", "hops", "countries"})
	f := func(v float64) string { return fmt.Sprintf("%.2f", v) }
	for _, r := range rows {
		_ = c.Write([]string{r.vantage, r.target, r.location, fmt.Sprint(r.runs), f(r.connect), f(r.tls), f(r.rtt), f(r.jitter), f(r.calls20), f(r.page20),
			f(r.down1M), f(r.down10M), f(r.up1M), fmt.Sprintf("%.4f", r.udpLoss), f(r.hops), r.countries})
	}
	c.Flush()
}

func num(v float64, decimals int) string {
	if math.IsNaN(v) {
		return "n/a"
	}
	return fmt.Sprintf("%.*f", decimals, v)
}

func percent(v float64) string {
	if math.IsNaN(v) {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", v*100)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "report:", err)
	os.Exit(1)
}
