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
	hopLine = regexp.MustCompile(`^\s*(\d+)\s+(.*)$`)
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
	r := &results.Route{Tool: tool}
	scan := bufio.NewScanner(bytes.NewReader(out))
	for scan.Scan() {
		m := hopLine.FindStringSubmatch(scan.Text())
		if m == nil {
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
		if hop.Address != "" {
			hop.Name = reverse(ctx, hop.Address)
			hop.Country = country(geo, hop.Address)
		}
		r.Hops = append(r.Hops, hop)
	}
	return r, nil
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
