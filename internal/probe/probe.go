// Package probe measures targets from where it runs, as a user there would reach them.
package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/dekka-cloud/latency-bench/internal/results"
)

// Target is one deployment to measure, as the targets file lists it.
type Target struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	// UDPPort is the target's UDP echo port; 7001 when unset, and 0 skips UDP for it.
	UDPPort *int `json:"udp_port,omitempty"`
}

// Options sets what a run does.
type Options struct {
	Vantage   results.Vantage
	Rounds    int  // fresh connections and page loads per target
	Insecure  bool // accept self-signed certificates
	UDP       bool
	Route     bool
	GeoIP     *maxminddb.Reader
	Progress  io.Writer
	Version   string
	Downloads []int64
	Uploads   []int64
	Chatty    []int // sequences of API calls to time
}

// Defaults are the measurements a study run makes.
func Defaults() Options {
	return Options{
		Rounds:    5,
		UDP:       true,
		Route:     true,
		Downloads: []int64{100 << 10, 1 << 20, 10 << 20},
		Uploads:   []int64{100 << 10, 1 << 20},
		Chatty:    []int{10, 20, 30},
	}
}

// Run measures every target in turn, so they never compete for the same connection.
func Run(ctx context.Context, targets []Target, o Options) results.Run {
	run := results.Run{Version: results.Version, Probe: o.Version, Started: time.Now().UTC(), Vantage: o.Vantage}
	for _, t := range targets {
		say(o, "Measuring %s (%s)\n", t.Name, t.URL)
		run.Targets = append(run.Targets, measure(ctx, t, o))
	}
	return run
}

func measure(ctx context.Context, t Target, o Options) results.Target {
	out := results.Target{Name: t.Name, URL: t.URL}
	base, err := url.Parse(t.URL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		out.Error = "the URL must be http or https with a host"
		return out
	}
	c := client{insecure: o.Insecure, timeout: 15 * time.Second}
	keep := c.http(c.transport())

	info, err := describe(ctx, keep, t.URL)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.Location, out.Provider = info.Location, info.Provider

	say(o, "  first connections\n")
	for range o.Rounds {
		out.Connect = append(out.Connect, c.firstConnection(ctx, t.URL))
	}
	say(o, "  round trips\n")
	if samples, err := warm(ctx, keep, t.URL, 30); err == nil {
		out.Warm = results.Summarize(samples)
	}
	say(o, "  calls one after another\n")
	for _, n := range o.Chatty {
		if ch, err := chatty(ctx, keep, t.URL, n); err == nil {
			out.Chatty = append(out.Chatty, ch)
		}
	}
	say(o, "  page loads\n")
	for range o.Rounds {
		for _, n := range []int{10, 20} {
			if p, err := c.page(ctx, t.URL, n); err == nil {
				out.Page = append(out.Page, p)
			}
		}
	}
	say(o, "  downloads and uploads\n")
	for _, size := range o.Downloads {
		out.Download = append(out.Download, download(ctx, keep, t.URL, size))
	}
	for _, size := range o.Uploads {
		out.Upload = append(out.Upload, upload(ctx, keep, t.URL, size))
	}
	// UDP and the route go to the address the HTTP connections reached, so every
	// measurement follows the same path even when a name has several addresses.
	host := base.Hostname()
	for _, c := range out.Connect {
		if h, _, err := net.SplitHostPort(c.Address); err == nil && c.Error == "" {
			host = h
			break
		}
	}
	if o.UDP {
		port := 7001
		if t.UDPPort != nil {
			port = *t.UDPPort
		}
		if port > 0 {
			say(o, "  UDP jitter and loss\n")
			if u, err := udp(ctx, net.JoinHostPort(host, strconv.Itoa(port)), 100, 20*time.Millisecond); err == nil {
				out.UDP = u
			} else {
				say(o, "  couldn't test UDP: %v\n", err)
			}
		}
	}
	if o.Route {
		say(o, "  route\n")
		if r, err := route(ctx, host, o.GeoIP); err == nil {
			out.Route = r
		} else {
			say(o, "  couldn't trace the route: %v\n", err)
		}
	}
	return out
}

type targetInfo struct {
	Location string `json:"location"`
	Provider string `json:"provider"`
}

func describe(ctx context.Context, h *http.Client, base string) (targetInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/info", nil)
	if err != nil {
		return targetInfo{}, err
	}
	resp, err := h.Do(req)
	if err != nil {
		return targetInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return targetInfo{}, fmt.Errorf("%s/info answered %s; is this a latency-bench target?", base, resp.Status)
	}
	var info targetInfo
	return info, json.NewDecoder(resp.Body).Decode(&info)
}

func say(o Options, format string, args ...any) {
	if o.Progress != nil {
		fmt.Fprintf(o.Progress, format, args...)
	}
}
