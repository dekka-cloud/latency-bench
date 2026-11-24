package target

import (
	"context"
	"errors"
	"log/slog"
	"net"
)

// maxProbePacket is the largest UDP packet echoed. Probe packets are tiny, and refusing
// larger ones keeps the echo from amplifying anything.
const maxProbePacket = 64

// ServeUDP echoes each small packet back to its sender until ctx ends.
func ServeUDP(ctx context.Context, addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	buf := make([]byte, 1500)
	for {
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			slog.Warn("udp read", "err", err)
			continue
		}
		if n > maxProbePacket {
			continue
		}
		if _, err := conn.WriteTo(buf[:n], from); err != nil {
			slog.Warn("udp write", "err", err)
		}
	}
}
