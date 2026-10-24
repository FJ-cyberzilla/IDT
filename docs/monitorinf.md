# Guardcat: Intelligent Monitoring, Circuit Breaking & Auto-Finetuning Watchdog (`internal/guardcat`)

## 1. Executive Summary

**Guardcat** is the autonomous resilience, health monitoring, circuit-breaking, and auto-finetuning watchdog module within **IDT (Intelligent DNS Tool)**. Operating as an active sentinel across all test executions, warm-up phases, and real-time telemetry redraws, Guardcat ensures that transient network partitions, socket stalls, memory high-water marks, and server EOF (End-of-File) drops never compromise application stability or skew benchmark accuracy.

Guardcat operates locally, non-invasively, and asynchronously, protecting the runtime environment without altering system-level DNS settings or network configurations.

---

## 2. Duties & Core Responsibilities

Guardcat’s architectural mandate encompasses six primary responsibilities:

### 2.1 Endpoint Health Lifecycle & Status Tracking
Guardcat continuously tracks the operational health of every DNS resolver and network endpoint (`EndpointState`), classifying them into three distinct operational states:
- **`HEALTHY`**: Fully operational resolver responding within expected RTT thresholds with zero recent packet loss or protocol errors.
- **`DEGRADED`**: Experiencing intermittent packet loss, rising jitter, or early warning threshold violations (e.g., 2 consecutive failures). Probe frequency is dynamically throttled, and secondary fallback paths are pre-warmed.
- **`UNHEALTHY`**: Exceeded configured failure thresholds or EOF/reset error limits. Automatically triggers circuit breaking to quarantine the dead resolver and divert traffic to healthy fallbacks.

### 2.2 Threshold-Based Circuit Breaking & EOF Recovery
Network anomalies such as TCP RST (Reset) packets, premature EOF drops, or proxy tampering are tracked per endpoint:
- **Failure Threshold**: Default 3 consecutive failures.
- **EOF Trigger Threshold**: Default 3 EOF/reset errors within a rolling time window.
- When an endpoint breaches these thresholds, Guardcat immediately quarantines the node (`StatusUnhealthy`), logs an audit event, and signals the runner/vault to bypass the endpoint.

### 2.3 Memory & Goroutine Pressure Watchdog (`CheckMemory`)
Operating within resource-constrained environments (such as Termux on Android or embedded Linux servers), Guardcat monitors heap allocations and goroutine counts:
- **Optimal**: Heap allocation below configured watermark (default 512 MB).
- **Warning**: Heap allocation exceeds 75% of watermark (`MONITORING_HIGH_WATERMARK`).
- **Critical**: Heap allocation exceeds maximum watermark. Guardcat automatically invokes `runtime.GC()` to reclaim memory, re-evaluates heap status, and issues actionable recommendations (e.g., advising application restart if memory pressure persists).

### 2.4 Pre-Flight Readiness Probing (`EnsureReadiness`)
Before executing complex multi-protocol benchmarks or time-diagnostic engines, Guardcat executes lightweight UDP and TCP socket probes against trust anchors (`1.1.1.1:53`, `8.8.8.8:53`) to verify that the local socket stack, routing tables, and interface sockets are fully initialized and ready.

### 2.5 Internet Connectivity Awareness (`CheckInternetConnectivity`)
Guardcat detects whether the host environment is online or trapped behind captive portals, offline Wi-Fi, or disconnected VPN tunnels. If offline, it engages local fallback mode and logs clear diagnostic warnings, preventing hang states during benchmark execution.

### 2.6 Actionable Recommendation Engine (`RecommendAction`)
Guardcat synthesizes telemetry from endpoint states, memory reports, and circuit breaker metrics into concise, human-readable operational advice for the SOC command center or CI/CD logging pipeline.

---

## 3. Architecture & Internal State Machine

```
               [Incoming DNS Query / Probe]
                            │
                            ▼
                { Guardcat Watchdog }
               ┌──────────┴──────────┐
               │                     │
       [Success Recorded]    [Failure / EOF Recorded]
               │                     │
               ▼                     ▼
      (Reset Failure/EOF)    (Increment Failure/EOF Count)
               │                     │
               ▼                     ▼
        [Status: HEALTHY]     [Check Thresholds]
                             ┌───────┴───────┐
                       (< 3 Failures)   (>= 3 Failures / EOF)
                             │               │
                             ▼               ▼
                    [Status: DEGRADED]  [Status: UNHEALTHY]
                                             │
                                             ▼
                                    [Quarantine & Bypass]
```

---

## 4. API & Integration Reference

- **`New(cfg Config) *Guardcat`**: Instantiates Guardcat with custom threshold policies.
- **`RegisterEndpoint(name, address, protocol string)`**: Registers or updates a target endpoint.
- **`RecordSuccess(address, protocol string)`**: Records a successful query and resets error counters.
- **`RecordFailure(address, protocol string, err error)`**: Records runtime errors, detects TCP/EOF drops, and triggers circuit breaking.
- **`CheckMemory() MemoryReport`**: Reads runtime memory stats and triggers automated GC on high pressure.
- **`EnsureReadiness(ctx context.Context, target string) error`**: Validates socket readiness.
- **`RecommendAction() string`**: Generates synthesized operator recommendations.
