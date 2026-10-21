# IDT (Intelligent DNS Tool & Enterprise Simulation Suite) 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Cobra CLI](https://img.shields.io/badge/CLI-Cobra-1B1F23?style=flat&logo=github)](https://github.com/spf13/cobra)
[![Bubble Tea TUI](https://img.shields.io/badge/TUI-Bubble%20Tea-FF5E5B?style=flat&logo=charm)](https://github.com/charmbracelet/bubbletea)
[![Test Coverage](https://img.shields.io/badge/Tests-100%25%20Passed-9ECE6A?style=flat&logo=checkmarx)](https://github.com/cyberzilla-systems/idt)
[![License: MIT](https://img.shields.io/badge/License-MIT-E0AF68.svg)](LICENSE)

**IDT (Intelligent DNS Tool)** is a professional, high-performance, enterprise-grade DNS benchmarking, diagnostics, tuning, and enterprise application simulation suite written in Go. Created by **FJ™ Cybertronic Systems**, it combines a robust headless CLI engine powered by **Cobra** with an interactive real-time terminal user interface (TUI) powered by **Bubble Tea** and **Lip Gloss**.

---

## 🎯 The Goal & Enterprise Scope

In modern network engineering and distributed enterprise architectures, validating application performance across remote offices, testing multi-protocol behavior (HTTP, VoIP/SIP-RTP, SaaS, custom microservices) under dynamic conditions (congestion, jitter, packet drop), and capturing/replaying production traffic in controlled lab settings is critical before production rollout. **IDT** provides network administrators, security engineers, and DevOps operators with an ultra-fast, zero-logging telemetry and simulation engine.

---

## 🔒 Core Capabilities, System DNS Discovery & Resilience

- **🌐 Automated Multi-Platform System DNS Discovery (`internal/config/system_dns.go`)**: Automatically detects and probes system DNS resolvers non-conflictually across Windows PowerShell (`Get-DnsClientServerAddress` / `ipconfig`), WSL 2 host bridges (`/mnt/c/`), Android system properties (`net.dns1`), Termux (`/etc/resolv.conf`), DoH, and DoT.
- **⚡ Token-Bucket Rate Limiter (`internal/helper/ratelimit.go`)**: Enterprise-grade rate limiting protecting upstream resolvers from 429 rate-limiting and IP flagging.
- **🔄 Automated Stall Reset & Retry Handling (`internal/runner/runner.go`)**: Intelligent network stall detection with automatic single retry and fallback reporting ("Not reachable this time") upon persistent failure.
- **📡 Multi-Protocol Dialer Suite (`internal/dialer/`)**: Supports DNS-over-HTTPS (DoH), DNS-over-TLS (DoT), Standard UDP/TCP, UDP Datagram Probes (`udp.go`), and custom Wire encoders/decoders.
- **🏢 Advanced Enterprise Simulation (`internal/enterprise/`)**:
  - **DNSSEC Audit (`dnssec.go`)**: Validates cryptographic trust chains and signature validity.
  - **Timezone Analytics (`timezone.go`)**: Analyzes temporal routing latency across global timezones.
  - **Remote Office Matrix (`matrix.go`)**: Simulates latency and SLA compliance across global data centers.
  - **Live Traffic Replay (`replay.go`)**: Replays production traffic capture logs with configurable speed multipliers.

---

## 📦 Installation & Usage

Ensure **Go 1.21+** is installed.

```bash
# Clone the repository
git clone https://github.com/cyberzilla-systems/idt.git
cd idt

# Run tests across all 15 packages
go test -count=1 ./...

# Build the binary
go build -o idt .

# Run the interactive TUI command center
./idt
```

---

## 📞 Support & Repository

- **Repository**: [https://github.com/cyberzilla-systems/idt](https://github.com/cyberzilla-systems/idt)
- **Support & Inquiries**: `cyberzilla.systsms@gmail.com`
- **Branding**: FJ™ Cybertronic Systems

---

## 📄 License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for more information.
