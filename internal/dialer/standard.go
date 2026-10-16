package dialer

import (
	"context"
	"net"
	"time"
)

type Standard struct {
	Address string
	Timeout time.Duration
}

func (s Standard) Dial(ctx context.Context, target string) (Answer, error) {
	d := net.Dialer{Timeout: s.Timeout}
	conn, err := d.DialContext(ctx, "udp", s.Address)
	if err != nil {
		return Answer{}, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(s.Timeout))

	q := buildQuery(target)
	start := time.Now()
	if _, err := conn.Write(q); err != nil {
		return Answer{}, err
	}
	resp := make([]byte, 512)
	n, err := conn.Read(resp)
	if err != nil {
		return Answer{}, err
	}
	rtt := time.Since(start)
	ips, _ := ParseA(resp[:n])
	return Answer{RTT: rtt, IPs: ips}, nil
}
