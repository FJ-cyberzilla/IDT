package errorhandling

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
)

// ErrorCategory classifies network and system errors for intelligent decision making.
type ErrorCategory string

const (
	CategoryTimeout     ErrorCategory = "TIMEOUT"
	CategoryDNS         ErrorCategory = "DNS_RESOLUTION_FAILURE"
	CategoryTLS         ErrorCategory = "TLS_HANDSHAKE_ERROR"
	CategoryRefused     ErrorCategory = "CONNECTION_REFUSED"
	CategoryUnreachable ErrorCategory = "NETWORK_UNREACHABLE"
	CategoryCanceled    ErrorCategory = "CANCELED"
	CategoryUnknown     ErrorCategory = "UNKNOWN"
)

// NetworkErrorDetails provides structured analysis and remediation advice for network test failures.
type NetworkErrorDetails struct {
	Category       ErrorCategory
	OriginalError  error
	Target         string
	ActionableHint string
	IsTransient    bool
	Timestamp      time.Time
}

// ResolveError inspects any error encountered during network testing (DNS, DoH, DoT, Standard dial)
// and returns rich diagnostic details, preventing silent crashes and providing exact user guidance.
func ResolveError(err error, target string) NetworkErrorDetails {
	now := time.Now()
	if err == nil {
		return NetworkErrorDetails{
			Category:       ErrorCategory("NONE"),
			Target:         target,
			ActionableHint: "Operation successful",
			IsTransient:    false,
			Timestamp:      now,
		}
	}

	errStr := err.Error()

	// Check for context canceled / deadline exceeded
	if errors.Is(err, context.Canceled) {
		return NetworkErrorDetails{
			Category:       CategoryCanceled,
			OriginalError:  err,
			Target:         target,
			ActionableHint: "Test was canceled by user or upstream context timeout.",
			IsTransient:    true,
			Timestamp:      now,
		}
	}
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "timeout") {
		return NetworkErrorDetails{
			Category:       CategoryTimeout,
			OriginalError:  err,
			Target:         target,
			ActionableHint: "Connection timed out. Check network latency, firewall state, or increase timeout settings.",
			IsTransient:    true,
			Timestamp:      now,
		}
	}

	// Check net.Error timeout or temporary
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return NetworkErrorDetails{
				Category:       CategoryTimeout,
				OriginalError:  err,
				Target:         target,
				ActionableHint: "Network timeout encountered. Target host might be throttling or offline.",
				IsTransient:    true,
				Timestamp:      now,
			}
		}
	}

	// DNS lookup errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) || strings.Contains(errStr, "no such host") || strings.Contains(errStr, "lookup") {
		category := CategoryDNS
		hint := "DNS resolution failed. Check Wi-Fi/cellular connection, Private DNS settings, or upstream DoH resolver."
		if IsAndroidPrivateDNSFailure(err) {
			category = CategoryPrivateDNS
			hint = "Android Private DNS enforcement active but VPN disconnected/unreachable. System rejected connections. Disable Private DNS in settings or connect VPN."
		}
		return NetworkErrorDetails{
			Category:       category,
			OriginalError:  err,
			Target:         target,
			ActionableHint: hint,
			IsTransient:    true,
			Timestamp:      now,
		}
	}

	// TLS handshake errors
	if strings.Contains(errStr, "tls:") || strings.Contains(errStr, "certificate") || strings.Contains(errStr, "handshake failure") {
		return NetworkErrorDetails{
			Category:       CategoryTLS,
			OriginalError:  err,
			Target:         target,
			ActionableHint: "TLS handshake error. Check device system clock, SSL inspection proxies, or root certificates.",
			IsTransient:    false,
			Timestamp:      now,
		}
	}

	// Connection refused (syscall ECONNREFUSED)
	if errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(errStr, "connection refused") {
		return NetworkErrorDetails{
			Category:       CategoryRefused,
			OriginalError:  err,
			Target:         target,
			ActionableHint: "Connection refused by target port. Service might be down or blocked by local firewall/ISP policy.",
			IsTransient:    false,
			Timestamp:      now,
		}
	}

	// Network unreachable (syscall ENETUNREACH)
	if errors.Is(err, syscall.ENETUNREACH) || strings.Contains(errStr, "network is unreachable") || strings.Contains(errStr, "no route to host") {
		return NetworkErrorDetails{
			Category:       CategoryUnreachable,
			OriginalError:  err,
			Target:         target,
			ActionableHint: "Network unreachable. Check airplane mode, Wi-Fi router, mobile data, or active VPN tunnel state.",
			IsTransient:    true,
			Timestamp:      now,
		}
	}

	// Fallback unknown
	return NetworkErrorDetails{
		Category:       CategoryUnknown,
		OriginalError:  err,
		Target:         target,
		ActionableHint: fmt.Sprintf("Unexpected network error: %v. Verify network interface connectivity.", err),
		IsTransient:    true,
		Timestamp:      now,
	}
}

// FormatDiagnostic returns a clean, human-readable summary string for terminal output or dashboard display.
func FormatDiagnostic(details NetworkErrorDetails) string {
	if details.Category == ErrorCategory("NONE") {
		return "✅ Status: OK"
	}
	return fmt.Sprintf("❌ [%s] Target: %s\n   Reason: %v\n   Advice: %s",
		details.Category, details.Target, details.OriginalError, details.ActionableHint)
}
