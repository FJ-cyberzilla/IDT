Advanced Coverage Improvement Plan — Go Codebase

Objective

Raise meaningful test coverage from 41.3% to a defensible ~75–80% by (a) testing what is testable with proper seams, (b) explicitly excluding what genuinely cannot be unit-tested, and (c) enforcing per-package coverage gates in CI. Coverage is treated as a diagnostic, not a vanity metric.

---

Phase 0 — Baseline & Governance (Week 1)

Files to add:

· .testcoverage.yml — per-package thresholds and exclusions
· .github/workflows/coverage.yml — CI gate using go-test-coverage
· docs/COVERAGE.md — documents which packages are excluded and why

Actions:

1. Run go test ./... -coverprofile=coverage.out -coverpkg=./... and record baseline per package.
2. Define exclusion list explicitly: main.go (main() body only), internal/dialer/standard.go (real net.Dial), internal/ui/theme.go (static constants), internal/ui/color.go (static constants).
3. Set per-package thresholds: configs 100%, vault 90%, dialer 85%, errorhandling 85%, enterprise 85%, compare 80%, helper 80%, stats 85%, config 85%, runner 80%, output 80%, tune 75%, check 70%, cmd 70%, ui 70%, compare 80%.
4. Fail CI if any listed package drops below its threshold.

Anti-pattern to enforce: ban assertion-free tests in review.

---

Phase 1 — UI Event Loop (Week 2)

Files to add:

· internal/ui/dashboard_test.go
· internal/ui/visual_test.go

Coverage targets: internal/ui/dashboard.go, internal/ui/visual.go, internal/ui/ui_test.go extended.

Actions:

1. Table-driven tests for every Update(msg) (Model, Cmd) branch — key messages, tea.WindowSizeMsg, custom messages, quit, resize, tab switches.
2. Assert on returned model fields and non-nil tea.Cmd where expected.
3. Golden-file tests for visual.go rendering using teatest with a fixed terminal size.
4. Test Init() returns expected commands.
5. Cover theme/color selection logic if non-constant.

---

Phase 2 — CLI Wiring (Week 2–3)

Files to add:

· internal/cmd/root_exec_test.go
· internal/cmd/flags_test.go

Refactor (minimal):

· Ensure NewRootCmd(deps) accepts dependencies as an interface (Runner, Checker, Vault).
· main.go only calls cmd.Execute() — no logic above main().

Actions:

1. Execute each subcommand via cmd.ExecuteC() with SetOut/SetErr buffers.
2. Assert stdout/stderr contents, exit codes, and error messages.
3. Test flag parsing, defaults, mutual exclusion, and required-flag validation.
4. Test --help, --version, unknown command, unknown flag paths.
5. Use stub dependencies for anything touching network or disk.

---

Phase 3 — Check Module (Week 3)

Files to add:

· internal/check/check_internal_test.go
· internal/check/fakes_test.go

Refactor:

· Introduce Resolver and Dialer interfaces inside check.go; constructor accepts them.

Actions:

1. Fake Resolver returning: valid response, NXDOMAIN, SERVFAIL, timeout, malformed packet.
2. Fake Dialer returning: success, connection refused, TLS handshake failure, timeout.
3. Table-driven tests for retry logic, fallback ordering, and per-server timeout accounting.
4. Assert aggregated results (success/failure counts, timing fields).
5. Only real net.Dial and real net.Resolver calls remain uncovered and are excluded in Phase 0.

---

Phase 4 — Runner, Output, Tune (Week 4)

Files to add:

· internal/runner/runner_concurrency_test.go
· internal/output/table_golden_test.go
· internal/tune/tune_test.go (extend)

Actions:

1. Runner: test worker pool with -race, cancellation via context, partial failures, ordered/unordered output.
2. Output: golden-file tests for table formatting across terminal widths and unicode edge cases.
3. Tune: test timeout/parallelism heuristics with table-driven inputs (RTT ranges, error rates).

---

Phase 5 — Vault & Config Hardening (Week 4–5)

Files to add:

· internal/vault/pool_concurrency_test.go
· internal/config/loader_edge_test.go

Actions:

1. Vault: concurrent encrypt/decrypt roundtrips under -race; corrupted .enc files in datalist/; wrong-key and truncated-ciphertext paths.
2. Config: malformed JSON, missing fields, unknown keys, env-var overrides, precedence order.
3. Add fuzz tests: FuzzCryptoRoundtrip, FuzzConfigLoader, FuzzResolverResponse — Go native fuzzing, run short corpus in CI.

---

Phase 6 — Integration & E2E Harness (Week 5)

Files to add:

· internal/testutil/fakes.go — shared fake implementations
· integration/ directory with build tag //go:build integration
· Makefile targets: test, test-race, test-integration, test-fuzz, coverage

Actions:

1. Shared fakes package prevents duplication across check/runner/cmd tests.
2. Integration tests (build-tagged, not run by default) cover real dialing to 127.0.0.1 listeners spun up in-process — covers standard.go without external network.
3. End-to-end CLI test: build binary, run against local listeners, assert exit codes and output.

---

Phase 7 — Fuzzing Expansion (Week 5–6)

Files to add:

· internal/dialer/parse_fuzz_test.go
· internal/dialer/wire_fuzz_test.go
· internal/errorhandling/resolve_fuzz_test.go

Actions:

1. Fuzz DNS packet parsing (parse.go, wire.go) — high-value, security-relevant.
2. Fuzz DoH/DoT response parsing.
3. Fuzz error classification inputs.
4. Seed corpora with real captures from datalist/.
5. Run fuzzers nightly in CI with a time budget.

---

Phase 8 — Documentation & Thresholds Lock (Week 6)

Files to add/update:

· docs/TESTING.md — expand with coverage philosophy, exclusion list, fuzz/integration instructions
· docs/COVERAGE.md — auto-generated coverage report committed on main
· README.md — badge showing coverage

Actions:

1. Document the "testable vs. excluded" boundary so future contributors know the rules.
2. Auto-generate coverage badge and per-package table on every merge to main.
3. Lock thresholds; future PRs must not lower any package.

---

Expected Outcome

Module Before Target
internal/ui low 70%
internal/cmd low 70%
internal/check low 70%
internal/runner — 80%
internal/output — 80%
internal/tune — 75%
internal/config — 85%
Overall 41.3% ~75–80%

---

Guardrails

1. No test exists solely to raise a number — every test asserts behavior.
2. Exclusions are explicit, documented, and reviewed.
3. -race is mandatory on all new concurrent tests.
4. Fuzz corpora are committed; regressions become unit tests.
5. Coverage gate is per-package, not global — prevents gaming.

Start with Phase 1 and Phase 2 (highest ROI, lowest risk), then Phase 3. Phases 4–7 are follow-on hardening; Phase 8 locks the gains.
