# Security Policy & Zero-Logging Guarantee

## 1. Overview
IDT (Intelligent DNS Tool) is designed from the ground up with a strict commitment to user privacy, cryptographic integrity, secure transport protocols, and zero telemetry collection.

## 2. Zero-Logging & Telemetry Policy
- **No Telemetry**: IDT does not collect, log, transmit, or store any user queries, IP addresses, diagnostic metrics, or session data.
- **Local Execution**: All benchmarking, DNS lookups, and statistical computations execute entirely locally on the operator's machine/device.
- **No External Callbacks**: IDT contains no analytics trackers, crash reporters, or external telemetry hooks.

## 3. Cryptographic Standards & Transport Security
- **TLS 1.3 Enforcement**: All DNS-over-HTTPS (DoH) and DNS-over-TLS (DoT) connections enforce strict TLS 1.3 or TLS 1.2 minimum standards with modern cipher suites.
- **Certificate Validation**: System root certificate stores are strictly validated during TLS handshakes. IDT never disables certificate verification (`InsecureSkipVerify` is strictly prohibited).
- **DNSSEC Validation**: Response integrity is validated through DNSSEC verification pipelines where supported by the resolver.

## 4. Network Safety & Connection Pooling
- **Connection Reuse**: IDT utilizes optimized connection pooling (`GetSharedTransport`) with HTTP/2 keep-alives and idle timeout bounds to prevent socket exhaustion and port leakage.
- **Private IP Protection**: Built-in IP validation routines (`IsPrivateIP`) prevent unintended routing anomalies across RFC 1918 and RFC 3927 address spaces.

## 5. Vulnerability Reporting
If you discover a security vulnerability within IDT, please report it responsibly by opening a private security advisory or contacting the maintainers directly. Do not disclose security vulnerabilities publicly until a patch has been released.
