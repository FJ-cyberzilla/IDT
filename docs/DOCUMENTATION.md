# IDT (Intelligent DNS Tool) - Comprehensive Architecture & Technical Specification

## 1. Executive Summary

**IDT (Intelligent DNS Tool)** is a professional, high-performance, enterprise-grade DNS benchmarking, diagnostics, tuning, and monitoring suite written in Go. It combines a robust headless CLI engine powered by **Cobra** with an interactive real-time terminal user interface (TUI) powered by **Bubble Tea** and **Lip Gloss**. 

IDT is engineered for low-latency network telemetry, strict privacy compliance, multi-protocol DNS resolution (DoH, DoT, Standard UDP/TCP), connection pooling optimization, pre-flight offline detection, advanced error resolution, and secure dynamic resolver pooling via the Vault instrumentation pipeline. IDT operates entirely locally and non-invasively; it never modifies, overrides, or affects Android, Termux, or system-wide DNS or network settings during tests, analysis, or execution.

---

## 2. Architecture & Project Layout

IDT adheres strictly to standard Go project layout conventions (`cmd/`, `internal/`, `configs/`, `docs/`), emphasizing clear separation of concerns, high modularity, zero circular dependencies, and complete test isolation.

```
/data/data/com.termux/files/home/IDT/
├── main.go                      # Single integrated entry point (delegates to Cobra root)
├── go.mod / go.sum              # Go module dependencies (Cobra, Lip Gloss, Bubble Tea)
├── datalist/                    # Seed CSV candidate resolver datasets by country (gitignored)
├── configs/
│   ├── default.go               # Production default master configuration module
│   └── default_test.go          # Unit tests for default configuration & JSON serialization
├── internal/
│   ├── check/                   # Health checking & policy compliance engine
│   ├── cmd/                     # Cobra root command center & TUI launcher
│   ├── compare/                 # Side-by-side DNS provider benchmarking & diffing
│   ├── config/                  # Configuration loader, servers catalog, & security policies
│   ├── dialer/                  # Multi-protocol DNS dialer stack (DoH, DoT, Standard, Wire, Parser)
│   ├── errorhandling/           # Granular error classification, fallback strategy, & remediation engine
│   ├── helper/                  # Advanced concurrency, connection pooling, retry & IP utilities
│   ├── output/                  # Formatted terminal output & ASCII table renderers
│   ├── runner/                  # Concurrent benchmark execution engine
│   ├── stats/                   # Statistical aggregation (percentiles, mean, stddev)
│   ├── tune/                    # Automated performance tuning & recommendation engine
│   ├── enterprise/              # Advanced enterprise simulation: office matrix, multi-protocol testing, & traffic replay
│   ├── ui/                      # Bubble Tea TUI state machine, Lip Gloss theme, & visual renderers
│   └── vault/                   # Vault package: candidate source loading, warm-up probing, oracle baselining, live pool, & JSONL instrumentation
└── docs/
    ├── DOCUMENTATION.md         # Full architectural & technical specification
    └── index.html               # Project documentation web portal
```

---

## 3. Core Entry Point & CLI Command Center (`main.go`, `internal/cmd/`)

### 3.1 Entry Point (`main.go`)
The root `main.go` acts as a thin wrapper ensuring clean bootstrap and panic recovery, immediately delegating execution to the Cobra command center:

```go
package main

import (
    "os"
    "github.com/your-org/idt/internal/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### 3.2 Cobra Command Structure (`internal/cmd/root.go`)
The Cobra root command coordinates all CLI flags and subcommands:
- **Global Flags**:
  - `--config`, `-c`: Path to custom JSON configuration file.
  - `--timeout`, `-t`: Request timeout duration (default: 5s).
  - `--concurrency`: Number of concurrent benchmark workers (default: 10).
  - `--json`: Output results in structured JSON format for automation/CI pipelines.
  - `--servers`: Custom list of DNS servers to query.
- **Subcommands / Modes**:
  - Default execution launches the interactive **Bubble Tea TUI Dashboard**.
  - Headless benchmark, check, tune, and compare modes for headless automation.

---

## 4. Interactive TUI & Visual Rendering Engine (`internal/ui/`)

IDT features a state-of-the-art terminal UI designed with **Lip Gloss** and **Bubble Tea**, providing a modern, responsive, and aesthetically striking experience.

### 4.1 Theme & Design System (`internal/ui/theme.go`)
- **Color Palette**:
  - *Midnight Blue* (`#1A1B26`) - Base background and card fills.
  - *Neon Cyan* (`#7DCFFF`) - Primary accent and active highlights.
  - *Emerald Green* (`#9ECE6A`) - Success indicators and optimal latency.
  - *Amber Gold* (`#E0AF68`) - Warning states and moderate performance.
  - *Crimson Red* (`#F7768E`) - Error alerts and offline warnings.
  - *Slate Gray* (`#565F89`) - Muted borders and secondary text.
- **Styling**: Rounded borders, adaptive padding, bold headers, and responsive margins that fit any terminal width (gracefully collapsing below 80 columns).

### 4.2 Interactive State Machine (`internal/ui/dashboard.go`)
- **Bubble Tea Model**: Manages application state including active view tab (`Benchmark`, `Check`, `Tune`, `Compare`, `Settings`), warm-up progress, spinner ticks, and keyboard inputs (`Tab`, `j`/`k`, `Up`/`Down`, `q`, `Ctrl+C`).
- **Animated Multi-Stage Warm-Up Phase**:
  - Stage 1: Network Interface & Gateway Discovery.
  - Stage 2: VPN / ISP Telemetry Inspection.
  - Stage 3: Root Certificate & Trust Store Verification.
  - Stage 4: Pre-flight DNS Resolution Reachability Test.
  - Each stage features real-time checkmarks (`[✔]`) and animated spinner frames.
- **Self-Aware Offline Detection**: Automatically detects network disconnection before execution, rendering a prominent Crimson Red Attention Card alerting the operator with fallback instructions.

### 4.3 Modular View Renderers (`internal/ui/visual.go`)
- **Benchmark View**: Real-time latency distribution graphs and server response tables.
- **Check View**: Policy compliance audits, security headers, and DNSSEC validation cards.
- **Tune View**: Automated bottleneck detection and latency optimization recommendations.
- **Compare View**: Side-by-side jitter and percentile comparison tables between primary and backup resolvers.
- **Settings View**: Active configuration parameters, timeout settings, and privacy scorecards.

---

## 5. Production Default Configuration (`configs/default.go`, `internal/config/`)

### 5.1 Default Configuration Master (`configs/default.go`)
IDT includes a robust production-grade master configuration module that ships with optimized defaults:
- **Application Parameters**: Default timeout (3s), concurrency (10), retry count (2), buffer sizes.
- **Curated Resolvers**:
  - **Cloudflare DoH**: `https://cloudflare-dns.com/dns-query` (Privacy Score: 9.8/10, Zero-Logging: Yes).
  - **Google DoH**: `https://dns.google/dns-query` (Privacy Score: 8.5/10, Zero-Logging: Yes).
  - **Quad9 DoT**: `tls://9.9.9.9:853` (Privacy Score: 9.9/10, Malware Blocking: Enabled).
  - **Cloudflare Standard**: `1.1.1.1:53` (High reliability, UDP/TCP).
- **Serialization**: Full JSON serialization (`MarshalJSON`, `UnmarshalJSON`) and file persistence support.

---

## 6. Network Dialers & DNS Transport Stack (`internal/dialer/`)

IDT supports multiple enterprise DNS protocols with a unified abstraction layer (`dialer.go`):

- **DNS-over-HTTPS (`doh.go`)**: Implements RFC 8484 compliant POST/GET requests over TLS 1.3 with HTTP/2 multiplexing.
- **DNS-over-TLS (`dot.go`)**: Implements RFC 7858 secure TCP connections over port 853 with session resumption and SNI enforcement.
- **Standard UDP/TCP (`standard.go`)**: Traditional raw socket resolution with automatic fallback from UDP to TCP on truncation (`TC bit`).
- **Wire Encoder/Decoder (`wire.go`, `parse.go`)**: Binary DNS packet builder and response parser handling DNS header flags, RCODEs, question sections, and resource records (A, AAAA, CNAME, TXT) with strict bounds checking.

---

## 7. Diagnostics, Benchmarking & Analytics (`internal/check`, `internal/runner`, `internal/tune`, `internal/compare`, `internal/stats`)

- **Runner (`runner.go`)**: Executes concurrent query batches across multiple dialers with precise microsecond timing.
- **Stats (`stats.go`)**: Computes statistical aggregations including minimum, maximum, arithmetic mean, median, p95/p99 percentiles, and standard deviation.
- **Check (`check.go`)**: Validates resolver integrity, DNSSEC compliance, and response authenticity against known expected IPs.
- **Tune (`tune.go`)**: Analyzes benchmark results to recommend optimal dialer parameters, connection pool sizes, and timeout thresholds.
- **Compare (`compare.go`)**: Evaluates performance differentials between multiple DNS providers under identical load conditions.

---

## 8. Advanced Helpers & Resilient Error Handling (`internal/helper`, `internal/errorhandling`)

### 8.1 Advanced Helpers (`internal/helper/advance_helper.go`)
- **`SafeRun` & `SafeGo`**: Panic-recovering concurrency primitives preventing background goroutine crashes.
- **Exponential Backoff with Jitter**: Robust retry wrapper for transient network failures.
- **Shared Transport Connection Pooling (`GetSharedTransport`)**: Configures optimized `http.Transport` with HTTP/2 keep-alives, idle connection pooling, and dial timeouts.
- **Private IP Detection (`IsPrivateIP`)**: RFC 1918 and RFC 3927 range checking for network interface discovery.

### 8.2 Error Resolution Engine (`internal/errorhandling/resolve.go`)
Classifies network and protocol errors into actionable categories (Timeout, Network Unreachable, TLS Handshake Failure, DNS SERVFAIL, Rate Limited) and provides precise remediation guidance for operators.

### 8.3 Robust Lookup & Fallback Strategy (`internal/errorhandling/fallback.go`)
Specifically engineered to handle Android Private DNS and disconnected VPN stalls:
- **`IsAndroidPrivateDNSFailure`**: Detects system connection rejections (`ENETUNREACH`, `ECONNREFUSED`, dial timeouts) caused by Android enforcing Private DNS when the VPN tunnel is offline.
- **`RobustLookupWithFallback`**: Bypasses stalled system stub resolvers by executing direct low-level UDP/TCP bootstrap queries to public resolvers with strict timeout bounds (1.5s - 2s), preventing command stalls and ensuring graceful error recovery and dynamic status reporting.

### 8.4 Automated System DNS Discovery (`internal/config/system_dns.go`)
- Automatically examines and tests whatever DNS is configured across operating systems non-conflictually:
  - **Windows PowerShell / `ipconfig`**: Parses active network adapter and client DNS server addresses via `ipconfig /all` and PowerShell (`Get-DnsClientServerAddress`).
  - **WSL 2**: Inspects WSL guest resolver (`/etc/resolv.conf`) and queries Windows host DNS bridge/PowerShell seamlessly.
  - **Android & Termux**: Inspects system properties (`net.dns1`, `net.dns2`, private DNS specifiers) and Termux `/etc/resolv.conf`.
  - **DoH & DoT**: Detects encrypted transport endpoints configured in environment or system settings.

---

## 9. Vault Instrumentation & Live Resolver Pool (`internal/vault/`)

The **Vault** subsystem manages candidate resolver seeds, warm-up probing, baseline comparison oracles, live pool management, and append-only JSONL instrumentation.

### 9.1 JSONL Instrumentation (`instrumentation.go`)
- Logs warm-up events, test events, and pool snapshots to a gitignored `runs/` directory as structured JSONL records.
- Operates through the `Logger` interface to ensure zero coupling with storage layers.

### 9.2 Candidate Source Loader & Crypto (`candidates.go`, `crypto.go`)
- Scans and parses country-specific encrypted seed files (`*.enc`) inside `datalist/`.
- Uses **AES-256-GCM** authenticated encryption with SHA-256 derived keys to secure candidate resolver lists at rest.
- Automatically decrypts encrypted seed files transparently at runtime during candidate loading.
- **Fresh Clone & Gitignore Handling**: Because `datalist/` is gitignored to keep repository sizes lean, a fresh clone from GitHub without `datalist/` will not crash or fail; instead, the Vault pool gracefully falls back to IDT's robust built-in hardcoded fallback chain (Cloudflare, Google, Quad9). Users can optionally place their own encrypted or plain CSV seed lists in a local `datalist/` directory to enable custom regional pool seeding.

### 9.3 Live Pool Manager (`pool.go`)
- **Warm-Up Stage**: Probes candidates strictly pinned over plain UDP port 53.
- **Baseline Oracle**: Compares results against an encrypted public oracle (Cloudflare DoH) to filter out intercepting or dead resolvers.
- **Fallback Chain**: Maintains hardcoded production fallbacks outside the CSV seed list.
- Picker & Background Refresh: Offers round-robin / weighted selection and 15-minute background timer refreshing with pool size alerts (< 3).

---

## 10. Advanced Enterprise & Protocol Simulation Suite (`internal/enterprise/`)

IDT features a dedicated enterprise simulation engine designed by **FJ™ Cybertronic Systems** to validate application performance and stability across distributed environments before rollout:

### 10.1 Remote Office & Data Center Performance Matrix (`matrix.go`)
- Concurrently simulates latency, jitter, packet loss, and throughput between distributed remote offices and data centers (New York, San Francisco, London, Frankfurt, Tokyo, Sydney).
- Computes SLA status (`OPTIMAL`, `WARNING`, `VIOLATED`) based on thresholds and packet drop metrics.

### 10.2 Multi-Protocol Dynamic Testing (`protocols.go`)
- Probes a variety of protocols (**HTTP/REST API**, **VoIP / SIP-RTP**, **SaaS / M365 / Salesforce**, **Custom App TCP/UDP**) across dynamic network conditions (Normal/Optimal, High Congestion, Extreme Jitter, Packet Loss/Drop).
- Calculates protocol health scores (0-100) and structured diagnostics.

### 10.3 Live Traffic Capture & Lab Replay Simulator (`replay.go`)
- Captures production traffic transactions and replays them in a controlled lab setting with speed multipliers (1.0x to 10.0x), jitter injection, and packet drop simulation.
- Automatically flags anomalies and high-latency spikes to ensure safe deployment rollouts.

---

## 11. Testing & Quality Assurance

IDT maintains **100% unit test coverage** across all internal and top-level packages:
- `configs/default_test.go`
- `internal/cmd/cmd_test.go`
- `internal/ui/ui_test.go`
- `internal/dialer/dialer_test.go`
- `internal/config/config_test.go`
- `internal/runner/runner_test.go`
- `internal/check/check_test.go`
- `internal/tune/tune_test.go`
- `internal/output/output_test.go`
- `internal/compare/compare_test.go`
- `internal/stats/stats_test.go`
- `internal/helper/helper_test.go`
- `internal/errorhandling/resolve_test.go`
- `internal/errorhandling/fallback_test.go`
- `internal/vault/instrumentation_test.go`
- `internal/vault/candidates_test.go`
- `internal/vault/pool_test.go`

Execution command:
```bash
go test -v ./...
```

---

## 12. Compilation & Usage

### 12.1 Build Executable
```bash
go build -o idt .
```

### 12.2 Run Interactive Dashboard
```bash
./idt
```

### 12.3 Run Headless Benchmark
```bash
./idt --concurrency 20 --timeout 3s
```
