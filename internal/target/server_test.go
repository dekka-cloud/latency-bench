package target

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func serve(t *testing.T, cfg Config) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(Handler(cfg))
	t.Cleanup(s.Close)
	return s
}

func TestEveryEndpointCanBeTimedFromAnotherOrigin(t *testing.T) {
	s := serve(t, Config{Location: "Accra, Ghana", Provider: "Dekka"})
	for _, path := range []string{"/info", "/ping", "/api/items/3", "/assets/1", "/page", "/bytes?n=10"} {
		resp, err := http.Get(s.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: %s", path, resp.Status)
		}
		for header, want := range map[string]string{"Access-Control-Allow-Origin": "*", "Timing-Allow-Origin": "*", "Cache-Control": "no-store"} {
			if got := resp.Header.Get(header); got != want {
				t.Errorf("%s: %s = %q, want %q", path, header, got, want)
			}
		}
	}
}

func TestInfoSaysWhereTheTargetRuns(t *testing.T) {
	s := serve(t, Config{Location: "Accra, Ghana", Provider: "Dekka", Version: "1.0.0"})
	resp, err := http.Get(s.URL + "/info")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var info map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	if info["location"] != "Accra, Ghana" || info["provider"] != "Dekka" || info["collects"] != false {
		t.Fatalf("info = %v", info)
	}
}

func TestTransfersAreExactSizesAndBounded(t *testing.T) {
	s := serve(t, Config{Location: "x", Provider: "y"})
	resp, err := http.Get(s.URL + "/bytes?n=3000000")
	if err != nil {
		t.Fatal(err)
	}
	n, _ := io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if n != 3000000 {
		t.Fatalf("downloaded %d bytes", n)
	}
	resp, err = http.Get(s.URL + "/bytes?n=999999999")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("an oversized download answered %s", resp.Status)
	}
	resp, err = http.Post(s.URL+"/upload", "application/octet-stream", bytes.NewReader(make([]byte, 123456)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var up map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&up); err != nil || up["bytes"] != 123456 {
		t.Fatalf("upload = %v, %v", up, err)
	}
}

func TestAPIRepliesAreTheSameSmallSize(t *testing.T) {
	s := serve(t, Config{Location: "x", Provider: "y"})
	for _, n := range []string{"1", "29"} {
		resp, err := http.Get(s.URL + "/api/items/" + n)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if len(body) != apiReply || !json.Valid(body) {
			t.Fatalf("item %s: %d bytes, valid JSON %v", n, len(body), json.Valid(body))
		}
	}
}

func TestSharedResultsAreKeptOnlyWhenCollecting(t *testing.T) {
	off := serve(t, Config{Location: "x", Provider: "y"})
	resp, err := http.Post(off.URL+"/results", "application/json", strings.NewReader(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("a target that doesn't collect answered %s", resp.Status)
	}

	dir := t.TempDir()
	on := serve(t, Config{Location: "x", Provider: "y", CollectDir: dir})
	for _, body := range []string{`{"vantage":{"city":"Accra"}}`, `not json`} {
		resp, err := http.Post(on.URL+"/results", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
	}
	data, err := os.ReadFile(filepath.Join(dir, "results.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 || !strings.Contains(lines[0], `"city":"Accra"`) {
		t.Fatalf("kept %q", data)
	}
}
