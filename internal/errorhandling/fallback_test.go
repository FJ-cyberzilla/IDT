package errorhandling

import (
	"context"
	"net"
	"syscall"
	"testing"
)

func TestIsAndroidPrivateDNSFailure(t *testing.T) {
	if IsAndroidPrivateDNSFailure(nil) {
		t.Error("expected false for nil error")
	}

	if !IsAndroidPrivateDNSFailure(syscall.ENETUNREACH) {
		t.Error("expected true for ENETUNREACH")
	}

	if !IsAndroidPrivateDNSFailure(syscall.ECONNREFUSED) {
		t.Error("expected true for ECONNREFUSED")
	}
}

func TestRobustLookupWithFallback_IPTarget(t *testing.T) {
	ctx := context.Background()
	res := RobustLookupWithFallback(ctx, "1.1.1.1")
	if len(res.IPs) != 1 || !res.IPs[0].Equal(net.ParseIP("1.1.1.1")) {
		t.Errorf("expected IP 1.1.1.1, got %v", res.IPs)
	}
	if res.UsedFallback {
		t.Error("expected UsedFallback to be false for direct IP")
	}
}

func TestRobustLookupWithFallback_Domain(t *testing.T) {
	ctx := context.Background()
	res := RobustLookupWithFallback(ctx, "example.com")
	if len(res.IPs) == 0 {
		t.Error("expected resolved IPs for example.com")
	}
}
