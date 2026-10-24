IDT
(Intelligent DNS Tool & Enterprise Simulation Suite)
Go Version
Cobra CLI
Bubble Tea TUI
Test Coverage
License: MIT
IDT (Intelligent DNS Tool) is a professional, high-performance, enterprise-grade DNS benchmarking, diagnostics, tuning, and enterprise application simulation suite written in Go. Created by FJ™ Cybertronic Systems, it combines a robust headless CLI engine powered by Cobra with an interactive real-time terminal user interface (TUI) powered by Bubble Tea and Lip Gloss.
🎯 The Goal & Enterprise Scope
In modern network engineering and distributed enterprise architectures, validating application performance across remote offices, testing multi-protocol behavior (HTTP, VoIP/SIP-RTP, SaaS, custom microservices) under dynamic conditions (congestion, jitter, packet drop), and capturing/replaying production traffic in controlled lab settings is critical before production rollout. IDT provides network administrators, security engineers, and DevOps operators with an ultra-fast, zero-logging telemetry and simulation engine.
🛠️ Interactive Management Suite (installation.sh)
IDT features an interactive management shell script (installation.sh) that automates setup, environment verification, building, global deployment, and testing across platforms (Termux, WSL2, Linux, macOS).
To launch the management suite:
chmod +x installation.sh
./installation.sh


Management Options Menu
====================================================
  FJ™ Cybertronic Systems // IDT Management Suite
====================================================
  SELECT MANAGEMENT OPERATION:
----------------------------------------------------
  1) Check Dependencies & Environment (Termux/WSL2/Linux/macOS)
  2) Build IDT Application Binary
  3) Install Globally (/usr/local/bin or Termux bin)
  4) Run Test Suite & Race Analyzer
  5) Update from Remote Git Repository
  6) Run IDT Command Center
  7) Exit Management Suite
----------------------------------------------------
  Enter choice [1-7]:


🔒 Core Capabilities, System DNS Discovery & Resilience
🌐 Automated Multi-Platform System DNS Discovery (internal/config/system_dns.go): Automatically detects and probes system DNS resolvers non-conflictually across Windows PowerShell (Get-DnsClientServerAddress / ipconfig), WSL 2 host bridges (/mnt/c/), Android system properties (net.dns1), Termux (/etc/resolv.conf), DoH, and DoT.
🛡️ DPI Bypass TLS Fragmentation (internal/dialer/dot.go): Advanced ClientHello SNI splitting with 1-byte leading fragment and TCP segmentation (TCP_NODELAY), bypassing deep packet inspection with automatic fallback to standard TLS dial.
🔄 DNS EOF Reconnect Detection (internal/runner/runner.go): Automated sliding-window telemetry (20s default window, configurable threshold) detecting persistent DNS EOF dropoffs and triggering proactive socket resets and reconnect recovery.
⚡ Token-Bucket Rate Limiter (internal/helper/ratelimit.go): Enterprise-grade rate limiting protecting upstream resolvers from 429 rate-limiting and IP flagging.
🔄 Automated Stall Reset & Retry Handling (internal/runner/runner.go): Intelligent network stall detection with automatic single retry and fallback reporting ("Not reachable this time") upon persistent failure.
📡 Multi-Protocol Dialer Suite (internal/dialer/): Supports DNS-over-HTTPS (DoH), DNS-over-TLS (DoT), Standard UDP/TCP, UDP Datagram Probes (udp.go), and custom Wire encoders/decoders.
🏢 Advanced Enterprise Simulation (internal/enterprise/):
DNSSEC Audit (dnssec.go): Validates cryptographic trust chains and signature validity.
Timezone Analytics (timezone.go): Analyzes temporal routing latency across global timezones.
Remote Office Matrix (matrix.go): Simulates latency and SLA compliance across global data centers.
Live Traffic Replay (replay.go): Replays production traffic capture logs with configurable speed multipliers.
📂 Project Architecture & Tree
.
├── LICENSE
├── README.md
├── bin
│   └── idt
├── configs
│   ├── default.go
│   ├── default_test.go
│   └── idt
├── coverage.out
├── datalist
│   ├── at.enc, ca.enc, ch.enc, de.enc, finland.enc
│   ├── gb.enc, ir.enc, jpn.enc, monoco.enc, swiss.enc, usa.enc
├── docs
│   ├── ARCHITECTURE.mmd
│   ├── CHANGELOG.md / .html
│   ├── COVERAGE.md / .html
│   ├── DOCUMENTATION.md / .html
│   ├── SECURITY.md / .html
│   ├── TESTING.md / .html
│   └── doc.css
├── installation.sh
├── internal
│   ├── check/        # System and network validation checks
│   ├── cmd/          # Cobra CLI commands & root execution
│   ├── compare/      # Resolver benchmark comparison engines
│   ├── config/       # Configuration loaders, policies, system DNS discovery
│   ├── dialer/       # DoH, DoT, UDP, TCP, wire protocol handlers
│   ├── enterprise/   # DNSSEC audit, matrix simulation, traffic replay
│   ├── errorhandling/# Fallback and error resolution logic
│   ├── guardcat/     # Security telemetry and guardrail checks
│   ├── helper/       # Rate limiters and utility functions
│   ├── output/       # Formatted tables and data presenters
│   ├── runner/       # Benchmark and probe execution loops
│   ├── stats/        # Statistical calculators and latency percentiles
│   ├── timediag/     # Temporal routing and time-series diagnostics
│   ├── tune/         # Network auto-tuning suites
│   ├── ui/           # Bubble Tea dashboard, Lip Gloss themes, components
│   └── vault/        # Encrypted credential pools & cryptographic instrumentation
├── main.go
├── robots.txt
├── scripts
│   └── md2html.py
└── servers.json


📦 Manual Installation & Usage
Ensure Go 1.21+ is installed on your system.
# Clone the repository
git clone https://github.com/cyberzilla-systems/idt.git
cd idt

# Run tests across all packages with race analyzer
go test -v -race -count=1 ./...

# Build the binary manually (automatically generates SHA-256 checksum manifest bin/idt.sha256)
go build -ldflags="-s -w" -o bin/idt main.go
sha256sum bin/idt > bin/idt.sha256
sha256sum -c bin/idt.sha256

# Run the interactive TUI command center
./idt

🔒 Security & Artifact Integrity
- **Authenticated Encryption (AES-GCM)** (`internal/vault/crypto.go`): Vault payloads (`datalist/*.enc`) use AES-256-GCM with PBKDF2 key derivation and built-in authentication tags. Any manual tampering with `.enc` files instantly triggers decryption failure.
- **Embedded HMAC Configuration Verification** (`internal/config/hmac.go`): Configuration and server files (`servers.json`) are verified on startup via HMAC-SHA256 signatures (`servers.json.sig`) before parsing server entries.
- **SHA-256 Checksum Manifests** (`bin/idt.sha256`): Automatically generated during local builds via `installation.sh` and GitHub Actions CI (`.github/workflows/go.yml`) to detect build tampering.


📞 Support & Repository
Repository: https://github.com/cyberzilla-systems/idt
Support & Inquiries: cyberzilla.systsms@gmail.com
Branding: FJ™ Cybertronic Systems
