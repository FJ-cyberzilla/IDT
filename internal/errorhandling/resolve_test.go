package errorhandling

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
)

func TestResolveError(t *testing.T) {
	// Nil error
	details := ResolveError(nil, "example.com")
	if details.Category != ErrorCategory("NONE") {
		t.Errorf("expected NONE category for nil error, got %s", details.Category)
	}
	diag := FormatDiagnostic(details)
	if diag == "" {
		t.Error("expected diagnostic string")
	}

	// Context canceled
	details = ResolveError(context.Canceled, "example.com")
	if details.Category != CategoryCanceled {
		t.Errorf("expected CANCELED category, got %s", details.Category)
	}

	// Timeout
	details = ResolveError(context.DeadlineExceeded, "example.com")
	if details.Category != CategoryTimeout {
		t.Errorf("expected TIMEOUT category, got %s", details.Category)
	}

	// DNS error
	dnsErr := &net.DNSError{Err: "no such host", Name: "bad.domain"}
	details = ResolveError(dnsErr, "bad.domain")
	if details.Category != CategoryDNS {
		t.Errorf("expected DNS category, got %s", details.Category)
	}

	// Connection refused
	details = ResolveError(syscall.ECONNREFUSED, "127.0.0.1:80")
	if details.Category != CategoryRefused {
		t.Errorf("expected REFUSED category, got %s", details.Category)
	}

	// Network unreachable
	details = ResolveError(syscall.ENETUNREACH, "8.8.8.8")
	if details.Category != CategoryUnreachable {
		t.Errorf("expected UNREACHABLE category, got %s", details.Category)
	}

	// TLS error
	tlsErr := errors.New("tls: bad certificate record")
	details = ResolveError(tlsErr, "secure.com")
	if details.Category != CategoryTLS {
		t.Errorf("expected TLS category, got %s", details.Category)
	}

	// Unknown error
	unkErr := errors.New("something went totally wrong")
	details = ResolveError(unkErr, "unknown.com")
	if details.Category != CategoryUnknown {
		t.Errorf("expected UNKNOWN category, got %s", details.Category)
	}

	diagUnk := FormatDiagnostic(details)
	if diagUnk == "" {
		t.Error("expected diagnostic string for unknown error")
	}
}
