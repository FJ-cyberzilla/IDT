package dialer

import (
	"context"
	"net"
	"time"
)

// UDP implements dedicated low-level raw UDP DNS queries (RFC 1035).
type UDP struct {
	Address string
	Timeout time.Duration
}

// Dial executes a raw UDP DNS lookup against the target address.
func (u UDP) Dial(ctx context.Context, target string) (Answer, error) {
	addr := u.Address
	if addr == "" {
		addr = "1.1.1.1:53"
	}
	timeout := u.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "udp", addr)
	if err != nil {
		return Answer{}, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return Answer{}, err
	}

	query := buildQuery(target)
	start := time.Now()
	if _, err := conn.Write(query); err != nil {
		return Answer{}, err
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return Answer{}, err
	}
	rtt := time.Since(start)

	ips, err := ParseA(buf[:n])
	if err != nil {
		return Answer{}, err
	}

	return Answer{
		RTT: rtt,
		IPs: ips,
	}, nil
}
