# Changelog

All notable changes to the **IDT** (Intelligent Diagnostic Toolkit) CLI application and enterprise platform are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.0.html).

---

## [Unreleased] - 2026-10-19

### Added
- **FJ™ Cybertronic Systems Branding**: Implemented vintage orange TrueColor exit branding (`#ff8700`) across command-line failure states and application exit banners.
- **TUI About View & Command Center Enhancements**: Added an interactive About Us view detailing FJ™ Cybertronic Systems specifications, version metadata, support contacts (`cyberzilla.systsms@gmail.com`), and repository links.
- **Advanced Enterprise Simulation Suite (`internal/enterprise`)**:
  - `matrix.go`: High-throughput enterprise topology simulation matrix.
  - `protocols.go`: Advanced protocol negotiation and stress-testing protocols.
  - `replay.go`: Session replay and diagnostic telemetry recorder.
  - Comprehensive unit tests (`enterprise_test.go`).
- **Responsive Mobile Terminal Adaptation**: Enhanced Bubble Tea TUI dashboard and visual renderers for seamless rendering on narrow/mobile terminal viewports (Termux, iOS/Android mobile terminals) with dynamic text truncation and padding.
- **Robust Navigation & Signal Safety**: Implemented graceful handling and guards for `ESC`, `Ctrl+C`, and `Ctrl+Z` signals across all TUI states (`stateDashboard`, `stateAbout`, `stateEnterpriseMenu`).
- **PBKDF2 Cryptographic Vault Security**: Upgraded vault passphrase derivation to robust PBKDF2 hashing standards.
- **GitHub Workflow Hardening**: Restricted CI workflow token permissions for enhanced supply-chain security.

### Changed
- Refactored `internal/ui/dashboard.go`, `internal/ui/visual.go`, and `internal/ui/theme.go` for modular separation of concerns.
- Optimized concurrency bounding and slice pre-allocation across network dialers and runners.

### Fixed
- Unhandled terminal resize panics and layout overflow on narrow viewports.
- Signal interception stability during long-running diagnostic runs.

---

## [1.0.0] - 2026-10-15

### Added
- Initial release of IDT (Intelligent Diagnostic Toolkit).
- Multi-protocol DNS & network diagnostic engine (`check`, `compare`, `dialer`, `runner`).
- Encrypted credential vault (`vault`).
- Charmbracelet Bubble Tea interactive TUI dashboard.
- Comprehensive test suite covering 15 Go packages.
