// Package target is the service deployed in each location under test. It answers the
// requests the probe and the browser page time, echoes UDP packets for jitter and loss,
// and can collect results that people choose to share.
package target

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config is read from the environment by cmd/target.
type Config struct {
	Location string // where this deployment runs, such as "Accra, Ghana"
	Provider string // who hosts it, such as "Dekka"
	Version  string
	// CollectDir, when set, keeps results people share from the browser page in
	// results.jsonl there. Nothing about the sender's address is kept.
	CollectDir string
	// TargetsFile, when set, is the list of targets the browser page measures, as JSON:
	// [{"name": "dekka-accra", "url": "https://accra.example.com"}, ...]. Without it the
	// page measures only the target that served it.
	TargetsFile string
	Web         fs.FS // the browser test page
}

const (
	maxTransfer = 50 << 20  // the largest download or upload a request may ask for
	maxShared   = 256 << 10 // the largest shared result
	assetSize   = 20 << 10  // one page asset, about the size of a small script or image
	apiReply    = 320       // bytes of JSON in one API reply, a typical small response
)

// payload is random bytes, so no compression along the path can shrink a transfer.
var payload = func() []byte {
	b := make([]byte, 1<<20)
	_, _ = rand.Read(b)
	return b
}()

// Handler serves every HTTP endpoint.
func Handler(cfg Config) http.Handler {
	mux := http.NewServeMux()
	collect := &collector{dir: cfg.CollectDir}

	mux.HandleFunc("GET /info", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"location": cfg.Location,
			"provider": cfg.Provider,
			"version":  cfg.Version,
			"collects": cfg.CollectDir != "",
			"time":     time.Now().UTC(),
		})
	})
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "ok")
	})
	mux.HandleFunc("GET /api/items/{n}", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.PathValue("n"))
		w.Header().Set("Content-Type", "application/json")
		// A small JSON object of a fixed size, standing in for one API answer.
		head := fmt.Sprintf(`{"id":%d,"name":"item %d","price":"GHS 25.00","in_stock":true,"tags":["sample","benchmark"],"description":"`, n, n)
		pad := max(apiReply-len(head)-len(`"}`), 0)
		_, _ = io.WriteString(w, head+strings.Repeat("x", pad)+`"}`)
	})
	mux.HandleFunc("GET /assets/{n}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(payload[:assetSize])
	})
	mux.HandleFunc("GET /page", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, page)
	})
	mux.HandleFunc("GET /bytes", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.ParseInt(r.URL.Query().Get("n"), 10, 64)
		if err != nil || n < 0 || n > maxTransfer {
			http.Error(w, fmt.Sprintf("n must be a byte count up to %d", maxTransfer), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
		for n > 0 {
			chunk := min(n, int64(len(payload)))
			if _, err := w.Write(payload[:chunk]); err != nil {
				return
			}
			n -= chunk
		}
	})
	mux.HandleFunc("POST /upload", func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		n, err := io.Copy(io.Discard, http.MaxBytesReader(w, r.Body, maxTransfer))
		if err != nil {
			http.Error(w, "the upload was cut short or too large", http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"bytes": n, "server_ms": float64(time.Since(started).Microseconds()) / 1000})
	})
	mux.HandleFunc("GET /targets.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if cfg.TargetsFile == "" {
			_, _ = io.WriteString(w, `[{"name":"`+cfg.Location+`","url":""}]`)
			return
		}
		data, err := os.ReadFile(cfg.TargetsFile)
		if err != nil {
			http.Error(w, "couldn't read the targets file", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("POST /results", collect.handle)
	if cfg.Web != nil {
		mux.Handle("GET /", http.FileServerFS(cfg.Web))
	}
	return headers(mux)
}

// headers lets any page time these endpoints: CORS for the requests, Timing-Allow-Origin
// so the browser exposes DNS, connect and TLS timings across origins, and no caching so
// every request really travels.
func headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type")
		h.Set("Timing-Allow-Origin", "*")
		h.Set("Cache-Control", "no-store")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// collector appends shared results to a file, one JSON object per line.
type collector struct {
	dir string
	mu  sync.Mutex
}

func (c *collector) handle(w http.ResponseWriter, r *http.Request) {
	if c.dir == "" {
		http.Error(w, "this target doesn't collect results", http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxShared))
	if err != nil {
		http.Error(w, "results must be JSON under 256 KB", http.StatusRequestEntityTooLarge)
		return
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "results must be a JSON object", http.StatusBadRequest)
		return
	}
	line, _ := json.Marshal(map[string]any{"received": time.Now().UTC(), "result": result})
	c.mu.Lock()
	defer c.mu.Unlock()
	f, err := os.OpenFile(filepath.Join(c.dir, "results.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		slog.Error("couldn't open the results file", "err", err)
		http.Error(w, "couldn't store the result", http.StatusInternalServerError)
		return
	}
	_, err = f.Write(append(line, '\n'))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		http.Error(w, "couldn't store the result", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// page is a small HTML document that names six assets, like a real page's scripts and images.
const page = `<!doctype html><html><head><meta charset="utf-8"><title>Benchmark page</title>
<link rel="stylesheet" href="/assets/1"><script src="/assets/2"></script></head>
<body><img src="/assets/3"><img src="/assets/4"><img src="/assets/5"><img src="/assets/6">
<p>A stand-in for an application page.</p></body></html>`
