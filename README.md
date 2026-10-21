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

## 🔒 Core Security, DNS Verification & Anti-Hijack Features

- **🛡️ Non-Leaking DNS Probing**: Validates Google DNS (`8.8.8.8`), Cloudflare DNS (`1.1.1.1`), DoH, and DoT endpoints securely with or without an active VPN tunnel (`tun0`) without disrupting or leaking system DNS settings.
- **🚨 Automated Hijacking & Tampering Detection**: Performs live cryptographic and semantic comparisons against trusted Cloudflare/Google DoH authoritative ground-truth records. Automatically flags local interception and DNS hijacking (`mismatch`).
- **⏱️ Real-Time Elapsed Time Tracking**: Features live elapsed time counters (`[MM:SS]`) during preflight warmup, connection pool allocation, and diagnostic/tuning runs.
- **📈 Jitter & Ping Interaction Tracking**: Computes composite stability scores (`Median + Jitter × 1.5 - SuccessBonus`) prioritizing stability over raw speed.

---

## 🏢 Advanced Enterprise & Protocol Simulation Modules (`internal/enterprise/`)

1. **🌐 Remote Office & Data Center Performance Matrix (`matrix.go`)**:
   - Concurrently simulates and measures latency, jitter, packet loss, and throughput across virtual enterprise remote offices and data centers (New York, San Francisco, London, Frankfurt, Tokyo, Sydney).
   - Computes SLA compliance status (`OPTIMAL`, `WARNING`, `VIOLATED`) with precision metrics.
2. **📡 Multi-Protocol Dynamic Testing (`protocols.go`)**:
   - Tests a variety of protocols (**HTTP/REST API**, **VoIP / SIP-RTP**, **SaaS / M365 / Salesforce**, **Custom App TCP/UDP**) across dynamic network conditions (Normal, High Congestion, Extreme Jitter, Packet Loss/Drop).
   - Computes protocol-specific health scores (0-100) and actionable diagnostics.
3. **📼 Live Traffic Capture & Lab Replay Simulator (`replay.go`)**:
   - Captures and replays live production traffic records in a controlled lab setting with configurable speed multipliers (1.0x - 10.0x), jitter injection, packet drop simulation, and automated anomaly detection.

---

## 📦 Installation & Usage

Ensure **Go 1.21+** is installed.

```bash
# Clone the repository
git clone https://github.com/cyberzilla-systems/idt.git
cd idt

# Run tests
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
