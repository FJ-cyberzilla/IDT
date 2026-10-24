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

// dpiFragConn wraps a net.Conn to perform TLS fragmentation (DPI Bypass).
// Splits the initial TLS ClientHello record into a tiny leading record (1 byte)
// and subsequent record segments over TCP with TCP_NODELAY, preventing DPI SNI inspection.
type dpiFragConn struct {
	net.Conn
	fragmented bool
}

func (f *dpiFragConn) Write(p []byte) (int, error) {
	if !f.fragmented && len(p) > 5 && p[0] == 0x16 && p[1] == 0x03 {
		f.fragmented = true
		recLen := binary.BigEndian.Uint16(p[3:5])
		if int(recLen) == len(p)-5 && recLen > 2 {
			// Record 1: 1 byte leading fragment payload
			rec1 := []byte{0x16, p[1], p[2], 0x00, 0x01, p[5]}

			// Record 2: remaining ClientHello payload
			remLen := recLen - 1
			rec2 := make([]byte, 5+remLen)
			rec2[0] = 0x16
			rec2[1] = p[1]
			rec2[2] = p[2]
			binary.BigEndian.PutUint16(rec2[3:5], remLen)
			copy(rec2[5:], p[6:])

			if tcpConn, ok := f.Conn.(*net.TCPConn); ok {
				tcpConn.SetNoDelay(true)
			}

			if _, err := f.Conn.Write(rec1); err != nil {
				return 0, err
			}
			time.Sleep(2 * time.Millisecond)
			if _, err := f.Conn.Write(rec2); err != nil {
				return 0, err
			}
			return len(p), nil
		}
	}
	return f.Conn.Write(p)
}

func (d DoT) Dial(ctx context.Context, target string) (Answer, error) {
	host, _, err := net.SplitHostPort(d.Address)
	if err != nil {
		host = d.Address
	}

	// 1. Attempt DPI Bypass TLS fragmentation dial first
	answer, err := d.dialWithFragmentation(ctx, host, target)
	if err == nil {
		return answer, nil
	}

	// 2. Fallback to standard TLS dial if server requires unfragmented connection
	return d.dialStandard(ctx, host, target)
}

func (d DoT) dialWithFragmentation(ctx context.Context, host, target string) (Answer, error) {
	dialer := &net.Dialer{Timeout: d.Timeout}
	rawConn, err := dialer.DialContext(ctx, "tcp", d.Address)
	if err != nil {
		return Answer{}, err
	}

	fragConn := &dpiFragConn{Conn: rawConn}
	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	tlsConn := tls.Client(fragConn, tlsConfig)
	if err := tlsConn.SetDeadline(time.Now().Add(d.Timeout)); err != nil {
		rawConn.Close()
		return Answer{}, err
	}

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		rawConn.Close()
		return Answer{}, err
	}

	return d.exchangeDNS(tlsConn, target)
}

func (d DoT) dialStandard(ctx context.Context, host, target string) (Answer, error) {
	dialer := &net.Dialer{Timeout: d.Timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", d.Address, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return Answer{}, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(d.Timeout)); err != nil {
		return Answer{}, err
	}

	return d.exchangeDNS(conn, target)
}

func (d DoT) exchangeDNS(conn net.Conn, target string) (Answer, error) {
	defer conn.Close()
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
