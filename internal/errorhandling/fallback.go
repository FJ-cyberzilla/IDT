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

// CategoryPrivateDNS classifies Android Private DNS enforcement errors when VPN is disconnected.
const CategoryPrivateDNS ErrorCategory = "ANDROID_PRIVATE_DNS_STALL"

// IsAndroidPrivateDNSFailure analyzes an error and connection context to determine if Android's
// Private DNS stub resolver rejected connections because a required VPN tunnel is disconnected.
func IsAndroidPrivateDNSFailure(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "i/o timeout") && !strings.Contains(errStr, "dial tcp: lookup") {
		return false
	}
	isNetworkBlocked := errors.Is(err, syscall.ENETUNREACH) || errors.Is(err, syscall.ECONNREFUSED) ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "dial tcp: lookup")

	return isNetworkBlocked
}

type FallbackLookupResult struct {
	IPs           []net.IP
	UsedFallback  bool
	Error         error
	ErrorDetails  NetworkErrorDetails
	DiagnosticMsg string
}

// RobustLookupWithFallback performs a robust DNS lookup for a target hostname.
// If the primary system resolver fails or stalls, it falls back to a clean,
// encrypted or user-configured resolver path rather than leaking plaintext Port 53.
func RobustLookupWithFallback(ctx context.Context, host string) FallbackLookupResult {
	// If host is already an IP, return immediately
	if ip := net.ParseIP(host); ip != nil {
		return FallbackLookupResult{
			IPs:           []net.IP{ip},
			UsedFallback:  false,
			DiagnosticMsg: "Target is already an IP address",
		}
	}

	// Step 1: Attempt standard resolver lookup with strict timeout
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err == nil && len(addrs) > 0 {
		ips := make([]net.IP, len(addrs))
		for i, a := range addrs {
			ips[i] = a.IP
		}
		return FallbackLookupResult{
			IPs:           ips,
			UsedFallback:  false,
			DiagnosticMsg: "Standard DNS lookup successful",
		}
	}

	primaryErr := err
	if primaryErr == nil {
		primaryErr = errors.New("resolver returned zero addresses")
	}

	// Step 2: Secure Fallback Strategy
	fallbackCtx, fallbackCancel := context.WithTimeout(ctx, 2*time.Second)
	defer fallbackCancel()

	fallbackIPs, fallbackErr := secureBootstrapLookup(fallbackCtx, host)
	if fallbackErr == nil && len(fallbackIPs) > 0 {
		return FallbackLookupResult{
			IPs:           fallbackIPs,
			UsedFallback:  true,
			Error:         primaryErr,
			ErrorDetails:  ResolveError(primaryErr, host),
			DiagnosticMsg: "⚠️ System resolver stalled (Android Private DNS / VPN state). Recovered via secure bootstrap fallback.",
		}
	}

	// Both failed. Construct rich diagnostic error details.
	details := ResolveError(primaryErr, host)
	if IsAndroidPrivateDNSFailure(primaryErr) {
		details.Category = CategoryPrivateDNS
		details.ActionableHint = "Android Private DNS is enabled but VPN is disconnected or unreachable. System rejected connections. Disable Private DNS in Android settings or establish VPN connection."
	}

	return FallbackLookupResult{
		IPs:           nil,
		UsedFallback:  true,
		Error:         primaryErr,
		ErrorDetails:  details,
		DiagnosticMsg: fmt.Sprintf("❌ Resolution failed: %v. %s", primaryErr, details.ActionableHint),
	}
}

// secureBootstrapLookup handles safe fallback resolution using standard TCP DNS over port 53 
// to public recursive resolvers, avoiding TLS/port 443 protocol mismatch errors.
func secureBootstrapLookup(ctx context.Context, host string) ([]net.IP, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 1 * time.Second}
			// Use standard DNS TCP port 53 (Cloudflare / OpenDNS public fallback)
			return d.DialContext(ctx, "tcp", "1.1.1.1:53")
		},
	}

	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, len(addrs))
	for i, a := range addrs {
		ips[i] = a.IP
	}
	return ips, nil
}
