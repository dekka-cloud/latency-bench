// Command target is the service latency-bench measures. Run one in every location you want
// to compare, usually with docker compose (see deploy/).
//
//	BENCH_LOCATION   where it runs, such as "Accra, Ghana"          (required)
//	BENCH_PROVIDER   who hosts it, such as "Dekka"                    (required)
//	BENCH_HTTP       plain HTTP address                               (default :8080)
//	BENCH_HTTPS      HTTPS address                                    (default :8443)
//	BENCH_UDP        UDP echo address                                 (default :7001)
//	BENCH_TLS_CERT   certificate and key files; a self-signed pair is made when unset
//	BENCH_TLS_KEY
//	BENCH_HOSTNAME   the name the self-signed certificate covers
//	BENCH_COLLECT    a directory to keep results people share from the browser page
//	BENCH_TARGETS    a JSON file listing the targets the browser page measures
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dekka-cloud/latency-bench/internal/target"
	"github.com/dekka-cloud/latency-bench/web"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("target stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := target.Config{
		Location:    os.Getenv("BENCH_LOCATION"),
		Provider:    os.Getenv("BENCH_PROVIDER"),
		Version:     version,
		CollectDir:  os.Getenv("BENCH_COLLECT"),
		TargetsFile: os.Getenv("BENCH_TARGETS"),
		Web:         web.Files,
	}
	if cfg.Location == "" || cfg.Provider == "" {
		return errors.New("set BENCH_LOCATION and BENCH_PROVIDER so results say where this target runs")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	handler := target.Handler(cfg)
	plain := server(env("BENCH_HTTP", ":8080"), handler)
	secure := server(env("BENCH_HTTPS", ":8443"), handler)
	cert, err := certificate()
	if err != nil {
		return err
	}
	secure.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}

	errs := make(chan error, 3)
	go func() { errs <- plain.ListenAndServe() }()
	go func() { errs <- secure.ListenAndServeTLS("", "") }()
	go func() { errs <- target.ServeUDP(ctx, env("BENCH_UDP", ":7001")) }()
	slog.Info("target running", "location", cfg.Location, "provider", cfg.Provider, "http", plain.Addr, "https", secure.Addr)

	select {
	case <-ctx.Done():
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = plain.Shutdown(shutdown)
	_ = secure.Shutdown(shutdown)
	return nil
}

func server(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func certificate() (tls.Certificate, error) {
	certFile, keyFile := os.Getenv("BENCH_TLS_CERT"), os.Getenv("BENCH_TLS_KEY")
	if certFile != "" || keyFile != "" {
		return tls.LoadX509KeyPair(certFile, keyFile)
	}
	hosts := []string{"localhost"}
	if h := strings.TrimSpace(os.Getenv("BENCH_HOSTNAME")); h != "" {
		hosts = append(hosts, h)
	}
	cert, fingerprint, err := target.SelfSigned(hosts)
	if err == nil {
		slog.Info("serving HTTPS with a self-signed certificate; probe it with --insecure", "sha256", fingerprint)
	}
	return cert, err
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
