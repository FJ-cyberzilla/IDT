# Test Coverage Policy & Architecture

## Objective

To ensure robust reliability, correctness, and regression prevention across the IDT codebase, we target a meaningful statement coverage of ~75–80%. Coverage is treated as a diagnostic metric to guide test writing, not as a vanity score.

## Excluded Files and Rationale

Certain parts of the codebase interact directly with operating system network primitives or static presentation constants where unit testing adds no meaningful signal or is infeasible in CI isolation:

1. **`main.go` (Entrypoint):** Only wires command execution (`cmd.Execute()`). Tested via CLI integration tests rather than unit testing `main()`.
2. **`internal/dialer/standard.go`:** Performs raw production network dialing (`net.Dial`) against live destinations. Tested via integration tests and fakes elsewhere.
3. **`internal/ui/theme.go` & `internal/ui/color.go`:** Pure static styling constants and ansi color code helpers.

## Per-Package Thresholds

| Package | Minimum Threshold |
|---|---|
| `configs` | 100% |
| `vault` | 80% |
| `dialer` | 70% |
| `errorhandling` | 80% |
| `enterprise` | 80% |
| `compare` | 80% |
| `helper` | 80% |
| `stats` | 80% |
| `config` | 80% |
| `runner` | 70% |
| `output` | 70% |
| `tune` | 70% |
| `check` | 60% |
| `cmd` | 60% |
| `ui` | 60% |

## Enforcing Quality

- All PRs are gated by automated test execution and coverage verification via `go-test-coverage`.
- The `-race` flag is mandatory for concurrent tests and data structure validation.
