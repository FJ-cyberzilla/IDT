package dialer

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"time"
)

type DoT struct {
	Address string
	Timeout time.Duration
}

func (d DoT) Dial(ctx context.Context, target string) (Answer, error) {
	host, _, _ := net.SplitHostPort(d.Address)
	dialer := &net.Dialer{Timeout: d.Timeout}

	conn, err := tls.DialWithDialer(dialer, "tcp", d.Address, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return Answer{}, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(d.Timeout))

	q := buildQuery(target)
	prefixed := make([]byte, 2+len(q))
	binary.BigEndian.PutUint16(prefixed[:2], uint16(len(q)))
	copy(prefixed[2:], q)

	start := time.Now()
	if _, err := conn.Write(prefixed); err != nil {
		return Answer{}, err
	}
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return Answer{}, err
	}
	size := binary.BigEndian.Uint16(hdr)
	body := make([]byte, size)
	if _, err := io.ReadFull(conn, body); err != nil {
		return Answer{}, err
	}
	rtt := time.Since(start)
	ips, _ := ParseA(body)
	return Answer{RTT: rtt, IPs: ips}, nil
}
