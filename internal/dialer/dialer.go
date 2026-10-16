package dialer

import (
	"context"
	"net"
	"time"
)

// Answer is a single dial result.
type Answer struct {
	RTT time.Duration
	IPs []net.IP
}

// Dialer is implemented by every protocol.
// One interface. Three implementations. Nothing else.
type Dialer interface {
	Dial(ctx context.Context, target string) (Answer, error)
}
