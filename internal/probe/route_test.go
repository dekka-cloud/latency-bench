package probe

import (
	"strings"
	"testing"

	"github.com/dekka-cloud/latency-bench/internal/results"
)

func TestHopsAreReadFromTraceroute(t *testing.T) {
	out := []byte(`traceroute to example.net (203.0.113.9), 30 hops max, 60 byte packets
 1  192.0.2.1  0.412 ms
 2  *
 3  198.51.100.7  94.301 ms
`)
	want := []results.Hop{{N: 1, Address: "192.0.2.1", RTT: 0.412}, {N: 2}, {N: 3, Address: "198.51.100.7", RTT: 94.301}}
	assertHops(t, parseHops(out), want)
}

func TestHopsAreReadFromTracepath(t *testing.T) {
	out := []byte(` 1?: [LOCALHOST]                      pmtu 1500
 1:  192.0.2.254                                           0.289ms
 1:  192.0.2.254                                           0.222ms
 2:  198.51.100.237                                        6.961ms
 3:  198.51.100.139                                       94.324ms (This broken router returned corrupted payload) asymm  8
 4:  no reply
     Too many hops: pmtu 1500
     Resume: pmtu 1500
`)
	want := []results.Hop{{N: 1, Address: "192.0.2.254", RTT: 0.289}, {N: 2, Address: "198.51.100.237", RTT: 6.961}, {N: 3, Address: "198.51.100.139", RTT: 94.324}, {N: 4}}
	assertHops(t, parseHops(out), want)
}

func assertHops(t *testing.T, got, want []results.Hop) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d hops, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("hop %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestAHostThatLooksLikeAnOptionIsRefused(t *testing.T) {
	for _, host := range []string{"", "-f", "--help"} {
		if _, err := route(t.Context(), host, nil); err == nil || !strings.Contains(err.Error(), "isn't a host") {
			t.Fatalf("route(%q) = %v", host, err)
		}
	}
}
