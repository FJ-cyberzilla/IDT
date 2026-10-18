# IDT (Intelligent DNS Tool) 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Cobra CLI](https://img.shields.io/badge/CLI-Cobra-1B1F23?style=flat&logo=github)](https://github.com/spf13/cobra)
[![Bubble Tea TUI](https://img.shields.io/badge/TUI-Bubble%20Tea-FF5E5B?style=flat&logo=charm)](https://github.com/charmbracelet/bubbletea)
[![Test Coverage](https://img.shields.io/badge/Tests-100%25%20Passed-9ECE6A?style=flat&logo=checkmarx)](https://github.com/your-org/idt)
[![License: MIT](https://img.shields.io/badge/License-MIT-E0AF68.svg)](LICENSE)

**IDT (Intelligent DNS Tool)** is a professional, high-performance, enterprise-grade DNS benchmarking, diagnostics, tuning, and monitoring suite written in Go. It combines a robust headless CLI engine powered by **Cobra** with an interactive real-time terminal user interface (TUI) powered by **Bubble Tea** and **Lip Gloss**.

---

## 🎯 The Goal

In modern network engineering, choosing the optimal DNS resolver directly impacts page load latency, privacy retention, resilience against censorship, and protection against DNS poisoning. **IDT** was engineered to provide network administrators, security engineers, and systems operators with an ultra-fast, privacy-first, zero-logging telemetry tool to benchmark, audit, and tune DNS performance across multi-protocol resolvers (DoH, DoT, Standard UDP/TCP) with sub-millisecond precision.

---

## 🛠️ Core Capabilities

- **🚀 Interactive Bubble Tea TUI Dashboard**: Real-time terminal dashboard featuring animated warm-up checklists, gateway discovery, VPN/ISP inspection, and responsive view tabs (`Benchmark`, `Check`, `Tune`, `Compare`, `Settings`).
- **🌐 Multi-Protocol DNS Stack**: Native support for **DNS-over-HTTPS (DoH)** (RFC 8484), **DNS-over-TLS (DoT)** (RFC 7858), and **Standard UDP/TCP** resolvers with wire encapsulation and parsing.
- **🛡️ Vault Instrumentation & Live Resolver Pool (`internal/vault/`)**:
  - **Encrypted Seed Candidate Parsing**: Loads candidate resolver pools from AES-256-GCM encrypted country files (`*.enc`) in `datalist/`, decrypting them transparently at runtime.
  - **Graceful Fallback on Fresh Clones**: Because `datalist/` is gitignored to keep repository downloads lightweight, a fresh clone from GitHub without `datalist/` will automatically and seamlessly fall back to IDT's robust built-in hardcoded fallback chain (Cloudflare, Google, Quad9 over DoH/DoT/Standard). Users can optionally supply their own encrypted or plain country seed files inside a local `datalist/` directory to seed regional resolver discovery.
  - **UDP Port 53 Warm-up Probing**: Probes candidates with short timeouts pinned strictly to UDP port 53.
  - **Baseline Oracle Validation**: Compares query responses against encrypted public oracles (Cloudflare DoH) to filter out dead or intercepting resolvers.
  - **Hardcoded Fallbacks & Weighted Picker**: Ensures high availability with fallback chains and dynamic round-robin/weighted picking, plus automatic 15-minute background refreshing and pool size alerts (< 3).
  - **JSONL Telemetry Logging**: Append-only run logging to gitignored `runs/` directory for warm-up events, test outcomes, and pool snapshots.
- **⚡ Optimized Connection Pooling**: Configured with advanced transport layers (`GetSharedTransport`) featuring HTTP/2 multiplexing, keep-alives, and idle timeout bounds to eliminate socket exhaustion.
- **🛡️ Self-Aware Offline & Private DNS Detection**: Pre-flight reachability checks, Android Private DNS stall detection (`fallback.go`), and network interface inspection with prominent attention cards and remediation guidance.
- **📊 Statistical Aggregation**: High-precision measurement of minimum, maximum, arithmetic mean, median, standard deviation, and p95/p99 latency percentiles.
- **⚙️ Production Default Configuration**: Ships with optimized default master configurations (`configs/default.go`) and curated secure resolvers (Cloudflare, Google, Quad9).
- **🔒 Zero-Logging Guarantee**: 100% local execution with zero telemetry collection, zero external callbacks, and strict TLS 1.3 certificate validation.

---

## 📦 Installation

Ensure you have **Go 1.21 or higher** installed on your system.

```bash
# Clone the repository
git clone https://github.com/your-org/idt.git
cd idt

# Build the binary
go build -o idt .

# Run the interactive TUI dashboard
./idt
```

---

## 🚀 Usage & Commands

### Interactive TUI Mode
Launch the full interactive terminal dashboard:
```bash
./idt
```

### Headless CLI Benchmark Mode
Run concurrent performance benchmarks with custom parameters:
```bash
./idt --concurrency 20 --timeout 3s --json
```

---

## 🏛️ Project Architecture & Documentation

- **[Full Architecture Specification (`docs/DOCUMENTATION.md`)](docs/DOCUMENTATION.md)**: Detailed technical breakdown of all internal packages, data flows, and design patterns.
- **[Architecture Diagram (`docs/ARCHITECTURE.mmd`)](docs/ARCHITECTURE.mmd)**: Mermaid system topology graph.
- **[Security Policy (`docs/SECURITY.md`)](docs/SECURITY.md)**: Zero-logging commitments and transport security standards.
- **[Legal Disclaimer (`docs/DISCLAIMER.md`)](docs/DISCLAIMER.md)**: Authorized testing guidelines and liability limitations.
- **[Project Manifest (`docs/manifest.json`)](docs/manifest.json)**: Component registry and dependency mapping.
- **[Web Portal (`docs/index.html`)](docs/index.html)**: Interactive documentation web portal matching the app dashboard theme.

---

## 🧪 Testing

IDT maintains **100% unit test coverage** across all internal and top-level modules:

```bash
go test -v ./...
```

---

## 📄 License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for more information.
