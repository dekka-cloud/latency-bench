package probe

import (
	"context"
	"net"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dekka-cloud/latency-bench/internal/target"
)

func TestARunMeasuresEverythingAgainstALocalTarget(t *testing.T) {
	s := httptest.NewTLSServer(target.Handler(target.Config{Location: "test", Provider: "local"}))
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := pc.LocalAddr().(*net.UDPAddr).Port
	_ = pc.Close()
	go func() { _ = target.ServeUDP(ctx, net.JoinHostPort("127.0.0.1", strconv.Itoa(port))) }()

	o := Defaults()
	o.Rounds, o.Insecure, o.Route = 2, true, false
	o.Downloads, o.Uploads = []int64{64 << 10}, []int64{64 << 10}
	run := Run(ctx, []Target{{Name: "local", URL: s.URL, UDPPort: &port}}, o)

	got := run.Targets[0]
	if got.Error != "" {
		t.Fatal(got.Error)
	}
	if got.Location != "test" || got.Provider != "local" {
		t.Fatalf("described as %q by %q", got.Location, got.Provider)
	}
	if len(got.Connect) != 2 || got.Connect[0].TLS <= 0 || got.Connect[0].Proto == "" {
		t.Fatalf("first connections %+v", got.Connect)
	}
	if got.Warm.N != 30 || len(got.Chatty) != 3 || len(got.Page) != 4 {
		t.Fatalf("warm %d, chatty %d, page %d", got.Warm.N, len(got.Chatty), len(got.Page))
	}
	if got.Download[0].Bytes != 64<<10 || got.Upload[0].Error != "" {
		t.Fatalf("transfers %+v %+v", got.Download, got.Upload)
	}
	if got.UDP == nil || got.UDP.Received != 100 || got.UDP.Loss != 0 {
		t.Fatalf("udp %+v", got.UDP)
	}
}
