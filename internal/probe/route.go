package probe

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"net/netip"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/dekka-cloud/latency-bench/internal/results"
)

var (
	// A hop's number starts the line: "  5  41.78.188.139  94.3 ms" from traceroute, and
	// " 5:  41.78.188.139  94.324ms" or " 1?: [LOCALHOST]" from tracepath.
	hopLine = regexp.MustCompile(`^\s*(\d+)\??:?\s+(.*)$`)
	ipField = regexp.MustCompile(`\b(\d{1,3}(?:\.\d{1,3}){3}|[0-9a-fA-F:]*:[0-9a-fA-F:]+)\b`)
	msField = regexp.MustCompile(`([\d.]+)\s*ms`)
)

// route traces the path to host with the system's own tool, without resolving names, then
// looks up each hop's name, and its country when a GeoIP database is open.
func route(ctx context.Context, host string, geo *maxminddb.Reader) (*results.Route, error) {
	tool, args := traceCommand(host)
	if tool == "" {
		return nil, errors.New("no traceroute, tracepath or tracert on this machine")
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, tool, args...).Output() //nolint:gosec // a fixed tool and the target's own host name
	if err != nil && len(out) == 0 {
		return nil, err
	}
	r := &results.Route{Tool: tool, Hops: parseHops(out)}
	for i := range r.Hops {
		if hop := &r.Hops[i]; hop.Address != "" {
			hop.Name = reverse(ctx, hop.Address)
			hop.Country = country(geo, hop.Address)
		}
	}
	return r, nil
}

// parseHops reads one hop per hop number from traceroute's or tracepath's output. tracepath
// can print a hop twice and starts with a line for the machine itself; the first line for a
// hop that names an address wins, and a hop that never answers is kept without one.
func parseHops(out []byte) []results.Hop {
	var hops []results.Hop
	at := map[int]int{}
	scan := bufio.NewScanner(bytes.NewReader(out))
	for scan.Scan() {
		m := hopLine.FindStringSubmatch(scan.Text())
		if m == nil || strings.Contains(m[2], "[LOCALHOST]") {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		hop := results.Hop{N: n}
		if ip := ipField.FindString(m[2]); ip != "" {
			if addr, err := netip.ParseAddr(ip); err == nil {
				hop.Address = addr.String()
			}
		}
		if ms := msField.FindStringSubmatch(m[2]); ms != nil {
			hop.RTT, _ = strconv.ParseFloat(ms[1], 64)
		}
		if i, seen := at[n]; seen {
			if hops[i].Address == "" && hop.Address != "" {
				hops[i] = hop
			}
			continue
		}
		at[n] = len(hops)
		hops = append(hops, hop)
	}
	return hops
}

func traceCommand(host string) (string, []string) {
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("tracert"); err == nil {
			return p, []string{"-d", "-h", "30", "-w", "2000", host}
		}
		return "", nil
	}
	if p, err := exec.LookPath("traceroute"); err == nil {
		return p, []string{"-n", "-q", "1", "-w", "2", "-m", "30", host}
	}
	if p, err := exec.LookPath("tracepath"); err == nil {
		return p, []string{"-n", "-m", "30", host}
	}
	return "", nil
}

func reverse(ctx context.Context, addr string) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	names, err := net.DefaultResolver.LookupAddr(ctx, addr)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func country(geo *maxminddb.Reader, addr string) string {
	if geo == nil {
		return ""
	}
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return ""
	}
	var rec struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
		CountryCode string `maxminddb:"country_code"` // DB-IP's lite databases
	}
	if err := geo.Lookup(ip).Decode(&rec); err != nil {
		return ""
	}
	if rec.Country.ISOCode != "" {
		return rec.Country.ISOCode
	}
	return rec.CountryCode
}
