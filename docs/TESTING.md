# IDT Testing Guide & Methodology

Welcome to the official **IDT** (Intelligent Diagnostic Toolkit) testing documentation. This guide outlines our testing philosophy, suite architecture, execution commands, and guidelines for maintaining our 100% test pass rate across all packages.

---

## 1. Testing Philosophy

IDT is an enterprise-grade diagnostic engine and interactive terminal application. Reliability, performance, and memory safety are paramount. Our testing standards require:
- **100% Package Test Pass Rate**: Zero regressions tolerated across all 15 core Go packages.
- **Unit & Integration Coverage**: Core modules (`dialer`, `vault`, `runner`, `enterprise`, `ui`, `check`) must include robust unit tests and race-condition validation.
- **TUI & Visual Validation**: Terminal user interface components, responsive mobile adaptations, and theme renderers are tested via deterministic unit tests in `internal/ui/ui_test.go`.

---

## 2. Test Execution Commands

### Running All Tests
To execute the entire test suite across all packages (disabling test caching to ensure fresh verification):
```bash
go test -count=1 ./...
```

### Running Tests with Race Detection
To run tests with the Go race detector enabled (recommended for concurrency-heavy components like `dialer`, `runner`, and `enterprise`):
```bash
go test -race -count=1 ./...
```

### Running Tests for a Specific Package
To focus on a specific package (e.g., the UI rendering package or enterprise simulation suite):
```bash
go test -v -count=1 ./internal/ui
go test -v -count=1 ./internal/enterprise
go test -v -count=1 ./internal/vault
```

---

## 3. Package Test Breakdown

| Package Path | Description | Key Test Focus Areas |
| :--- | :--- | :--- |
| `idt/internal/check` | Diagnostic check engine | Protocol connectivity, timeout handling |
| `idt/internal/compare` | Multi-target latency/metric comparison | Sorting, metric delta calculations |
| `idt/internal/config` | Configuration and policy loader | JSON parsing, policy validation |
| `idt/internal/dialer` | Multi-protocol network dialer (DoH, DoT, Standard) | Connection establishment, fallback resolution |
| `idt/internal/enterprise` | Enterprise simulation suite (Matrix, Protocols, Replay) | High-throughput stress tests, replay buffering |
| `idt/internal/errorhandling` | Fallback and error resolution engine | Error classification, retry backoff |
| `idt/internal/runner` | Diagnostic execution orchestrator | Concurrency bounding, cancellation handling |
| `idt/internal/stats` | Statistical aggregation engine | Mean, median, standard deviation accuracy |
| `idt/internal/ui` | Charmbracelet Bubble Tea TUI & visual renderers | View rendering, responsive padding, navigation states |
| `idt/internal/vault` | Encrypted credential vault & PBKDF2 crypto | Encryption/decryption roundtrips, key derivation |

---

## 4. Writing New Tests

When contributing new features or bug fixes, you **must** include corresponding unit tests:
1. **Naming Convention**: Test files must be named `*_test.go` and reside in the same package as the code under test.
2. **Table-Driven Tests**: Prefer idiomatic Go table-driven tests (`tests := []struct{...}`) to validate multiple input combinations cleanly.
3. **Assertions**: Use standard library `t.Errorf`, `t.Fatalf`, or `t.Logf` with clear diagnostic messages.
4. **Cleanup**: Always clean up temporary files or network listeners using `t.Cleanup()` or `defer`.

---

## 5. Continuous Integration (CI)

Our GitHub Actions workflow (`.github/workflows/go.yml`) automatically executes:
- Dependency vetting and module download.
- Static analysis and compilation across multiple Go versions.
- Full test suite execution (`go test -count=1 ./...`) with strict permission boundaries.
