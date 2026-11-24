package probe

import (
	"context"
	"encoding/binary"
	"net"
	"sync/atomic"
	"time"

	"github.com/dekka-cloud/latency-bench/internal/results"
)

// udp sends count small packets at a steady interval, as a voice call does, and times each
// echo. A packet that isn't back within a second counts as lost.
func udp(ctx context.Context, addr string, count int, interval time.Duration) (*results.UDP, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "udp", addr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	sent := make([]atomic.Int64, count) // when each packet left, in Unix nanoseconds
	rtts := make([]float64, 0, count)
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 64)
		seen := make(map[uint32]bool, count)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(time.Second))
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			if n < 4 {
				continue
			}
			seq := binary.BigEndian.Uint32(buf)
			if int(seq) >= count || seen[seq] {
				continue
			}
			left := sent[seq].Load()
			if left == 0 {
				continue
			}
			seen[seq] = true
			rtts = append(rtts, float64(time.Now().UnixNano()-left)/1e6)
			if len(seen) == count {
				return
			}
		}
	}()

	packet := make([]byte, 16)
	for i := range count {
		binary.BigEndian.PutUint32(packet, uint32(i))
		sent[i].Store(time.Now().UnixNano())
		if _, err := conn.Write(packet); err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
	<-done
	out := &results.UDP{Sent: count, Received: len(rtts), RTT: results.Summarize(rtts)}
	out.Loss = float64(count-len(rtts)) / float64(count)
	return out, nil
}
