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
        CategoryNone        ErrorCategory = "NONE"
        CategoryTimeout     ErrorCategory = "TIMEOUT"
        CategoryDNS         ErrorCategory = "DNS_RESOLUTION_FAILURE"
        CategoryTLS         ErrorCategory = "TLS_HANDSHAKE_ERROR"
        CategoryRefused     ErrorCategory = "CONNECTION_REFUSED"
        CategoryUnreachable ErrorCategory = "NETWORK_UNREACHABLE"
        CategoryCanceled    ErrorCategory = "CANCELED"
        CategoryUnknown     ErrorCategory = "UNKNOWN"
)

// NetworkErrorDetails provides structured analysis and remediation advice for network test failures.
// Tagged for JSON export readiness to feed TUI dashboards or logging pipelines.
type NetworkErrorDetails struct {
        Category       ErrorCategory `json:"category"`
        OriginalError  string        `json:"original_error"`
        Target         string        `json:"target"`
        ActionableHint string        `json:"actionable_hint"`
        IsTransient    bool          `json:"is_transient"`
        Timestamp      time.Time     `json:"timestamp"`
}

// String implements fmt.Stringer for clean logging output.
func (d NetworkErrorDetails) String() string {
        if d.Category == CategoryNone {
                return "✅ Status: OK"
        }
        return fmt.Sprintf("❌ [%s] Target: %s\n   Reason: %s\n   Advice: %s",
                d.Category, d.Target, d.OriginalError, d.ActionableHint)
}

// ResolveError inspects any error encountered during network testing (DNS, DoH, DoT, Standard dial)
// and returns rich diagnostic details, preventing silent crashes and providing exact user guidance.
func ResolveError(err error, target string) NetworkErrorDetails {
        now := time.Now()
        if err == nil {
                return NetworkErrorDetails{
                        Category:       CategoryNone,
                        Target:         target,
                        ActionableHint: "Operation successful",
                        IsTransient:    false,
                        Timestamp:      now,
                }
        }

        errStr := err.Error()

        // 1. Context Canceled / Deadline Exceeded
        switch {
        case errors.Is(err, context.Canceled):
                return NetworkErrorDetails{
                        Category:       CategoryCanceled,
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: "Test was canceled by user or upstream context timeout.",
                        IsTransient:    true,
                        Timestamp:      now,
                }
        case errors.Is(err, context.DeadlineExceeded) || strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "timeout"):
                return NetworkErrorDetails{
                        Category:       CategoryTimeout,
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: "Connection timed out. Check network latency, firewall state, or increase timeout settings.",
                        IsTransient:    true,
                        Timestamp:      now,
                }
        }

        // 2. Net.Error timeout or temporary interface check
        var netErr net.Error
        if errors.As(err, &netErr) && netErr.Timeout() {
                return NetworkErrorDetails{
                        Category:       CategoryTimeout,
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: "Network timeout encountered. Target host might be throttling or offline.",
                        IsTransient:    true,
                        Timestamp:      now,
                }
        }

        // 3. DNS Lookup Failures & Android Private DNS Handlers
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
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: hint,
                        IsTransient:    true,
                        Timestamp:      now,
                }
        }

        // 4. TLS Handshake & Certificate Failures
        if strings.Contains(errStr, "tls:") || strings.Contains(errStr, "certificate") || strings.Contains(errStr, "handshake failure") {
                return NetworkErrorDetails{
                        Category:       CategoryTLS,
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: "TLS handshake error. Check device system clock, SSL inspection proxies, or root certificates.",
                        IsTransient:    false,
                        Timestamp:      now,
                }
        }

        // 5. Connection Refused (ECONNREFUSED)
        if errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(errStr, "connection refused") {
                return NetworkErrorDetails{
                        Category:       CategoryRefused,
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: "Connection refused by target port. Service might be down or blocked by local firewall/ISP policy.",
                        IsTransient:    false,
                        Timestamp:      now,
                }
        }

        // 6. Network Unreachable (ENETUNREACH)
        if errors.Is(err, syscall.ENETUNREACH) || strings.Contains(errStr, "network is unreachable") || strings.Contains(errStr, "no route to host") {
                return NetworkErrorDetails{
                        Category:       CategoryUnreachable,
                        OriginalError:  errStr,
                        Target:         target,
                        ActionableHint: "Network unreachable. Check airplane mode, Wi-Fi router, mobile data, or active VPN tunnel state.",
                        IsTransient:    true,
                        Timestamp:      now,
                }
        }

        // 7. Unknown / Fallback Category
        return NetworkErrorDetails{
                Category:       CategoryUnknown,
                OriginalError:  errStr,
                Target:         target,
                ActionableHint: fmt.Sprintf("Unexpected network error: %v. Verify network interface connectivity.", err),
                IsTransient:    true,
                Timestamp:      now,
        }
}

// FormatDiagnostic returns a clean, human-readable summary string for terminal output or dashboard display.
// (Maintained for backward compatibility with existing code calling FormatDiagnostic directly).
func FormatDiagnostic(details NetworkErrorDetails) string {
        return details.String()
}
