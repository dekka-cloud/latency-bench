package probe

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/dekka-cloud/latency-bench/internal/results"
)

// client makes the HTTP clients a run needs: a fresh one per first connection, and one
// kept-alive client for everything that should reuse a connection.
type client struct {
	insecure bool
	timeout  time.Duration
}

func (c client) transport() *http.Transport {
	return &http.Transport{
		Proxy:               nil, // measure the path to the target, never a proxy's
		DialContext:         (&net.Dialer{Timeout: c.timeout}).DialContext,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: c.insecure, MinVersion: tls.VersionTLS12}, //nolint:gosec // only with --insecure, for self-signed targets
		ForceAttemptHTTP2:   true,
		MaxIdleConnsPerHost: 8,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}
}

func (c client) http(t *http.Transport) *http.Client {
	return &http.Client{Transport: t, Timeout: 2 * time.Minute}
}

// firstConnection times one request on a brand new connection, from looking up the name
// to the last byte of a small reply.
func (c client) firstConnection(ctx context.Context, base string) results.Connect {
	t := c.transport()
	defer t.CloseIdleConnections()
	var out results.Connect
	var dnsStart, connStart, tlsStart, wrote time.Time
	var mu sync.Mutex
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(httptrace.DNSDoneInfo) { out.DNS = since(dnsStart) },
		ConnectStart: func(string, string) {
			mu.Lock()
			if connStart.IsZero() {
				connStart = time.Now()
			}
			mu.Unlock()
		},
		ConnectDone: func(_, addr string, err error) {
			if err == nil {
				out.TCP = since(connStart)
				out.Address = addr
			}
		},
		TLSHandshakeStart:    func() { tlsStart = time.Now() },
		TLSHandshakeDone:     func(tls.ConnectionState, error) { out.TLS = since(tlsStart) },
		WroteRequest:         func(httptrace.WroteRequestInfo) { wrote = time.Now() },
		GotFirstResponseByte: func() { out.TTFB = since(wrote) },
	}
	started := time.Now()
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, base+"/ping", nil)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	resp, err := c.http(t).Do(req)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	out.Total = since(started)
	out.Proto = resp.Proto
	if resp.StatusCode != http.StatusOK {
		out.Error = resp.Status
	}
	return out
}

// get fetches path on h and reads the whole reply.
func get(ctx context.Context, h *http.Client, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := h.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	n, err := io.Copy(io.Discard, resp.Body)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s answered %s", url, resp.Status)
	}
	return n, err
}

// warm times small requests one after another on one open connection.
func warm(ctx context.Context, h *http.Client, base string, n int) ([]float64, error) {
	if _, err := get(ctx, h, base+"/ping"); err != nil {
		return nil, err
	}
	samples := make([]float64, 0, n)
	for i := range n {
		started := time.Now()
		if _, err := get(ctx, h, fmt.Sprintf("%s/ping?n=%d", base, i)); err != nil {
			return samples, err
		}
		samples = append(samples, since(started))
	}
	return samples, nil
}

// chatty makes calls API calls one after another, each waiting for the last.
func chatty(ctx context.Context, h *http.Client, base string, calls int) (results.Chatty, error) {
	started := time.Now()
	for i := range calls {
		if _, err := get(ctx, h, fmt.Sprintf("%s/api/items/%d", base, i)); err != nil {
			return results.Chatty{}, err
		}
	}
	return results.Chatty{Calls: calls, Total: since(started)}, nil
}

// page loads a page on a new connection: the HTML, its six assets in parallel, then calls
// API calls one after another, as a single-page app does after it starts.
func (c client) page(ctx context.Context, base string, calls int) (results.Page, error) {
	t := c.transport()
	defer t.CloseIdleConnections()
	h := c.http(t)
	started := time.Now()
	if _, err := get(ctx, h, base+"/page"); err != nil {
		return results.Page{}, err
	}
	html := since(started)
	var wg sync.WaitGroup
	errs := make(chan error, 6)
	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := get(ctx, h, fmt.Sprintf("%s/assets/%d", base, i)); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	if err := <-errs; err != nil {
		return results.Page{}, err
	}
	for i := range calls {
		if _, err := get(ctx, h, fmt.Sprintf("%s/api/items/%d", base, i)); err != nil {
			return results.Page{}, err
		}
	}
	return results.Page{Calls: calls, Total: since(started), HTML: html}, nil
}

func download(ctx context.Context, h *http.Client, base string, size int64) results.Transfer {
	started := time.Now()
	n, err := get(ctx, h, fmt.Sprintf("%s/bytes?n=%d", base, size))
	out := results.Transfer{Bytes: n, Time: since(started)}
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.Mbps = mbps(n, out.Time)
	return out
}

func upload(ctx context.Context, h *http.Client, base string, size int64) results.Transfer {
	body := make([]byte, size)
	_, _ = rand.Read(body)
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/upload", bytes.NewReader(body))
	if err != nil {
		return results.Transfer{Bytes: size, Error: err.Error()}
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := h.Do(req)
	if err != nil {
		return results.Transfer{Bytes: size, Error: err.Error()}
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	out := results.Transfer{Bytes: size, Time: since(started)}
	if resp.StatusCode != http.StatusOK {
		out.Error = resp.Status
		return out
	}
	out.Mbps = mbps(size, out.Time)
	return out
}

func since(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000
}

func mbps(bytes int64, ms float64) float64 {
	if ms <= 0 {
		return 0
	}
	return float64(bytes) * 8 / (ms * 1000)
}
