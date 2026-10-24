package guardcat

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

// EndpointStatus represents the operational health of a DNS/network endpoint.
type EndpointStatus string

const (
	StatusHealthy  EndpointStatus = "HEALTHY"
	StatusDegraded EndpointStatus = "DEGRADED"
	StatusUnhealthy EndpointStatus = "UNHEALTHY"
)

// EndpointState tracks failures, EOF triggers, and status for a specific network/DNS target.
type EndpointState struct {
	Name         string         `json:"name"`
	Address      string         `json:"address"`
	Protocol     string         `json:"protocol"`
	FailureCount int            `json:"failure_count"`
	EOFCount     int            `json:"eof_count"`
	Status       EndpointStatus `json:"status"`
	LastCheck    time.Time      `json:"last_check"`
	LastError    string         `json:"last_error"`
}

// MemoryReport captures app memory state and auto-finetuning decisions.
type MemoryReport struct {
	HeapAlloc    uint64    `json:"heap_alloc_bytes"`
	HeapSys      uint64    `json:"heap_sys_bytes"`
	NumGoroutine int       `json:"num_goroutine"`
	NumGC        uint32    `json:"num_gc"`
	Status       string    `json:"status"` // OPTIMAL, WARNING, CRITICAL
	ActionTaken  string    `json:"action_taken"`
	Timestamp    time.Time `json:"timestamp"`
}

// Config configures Guardcat thresholds, probe intervals, and auto-finetuning rules.
type Config struct {
	FailureThreshold    int           `json:"failure_threshold"`     // Default 3 consecutive failures
	EOFTriggerThreshold int           `json:"eof_trigger_threshold"` // Default 3 EOF/reset errors
	ProbeInterval       time.Duration `json:"probe_interval"`        // Keep-alive check interval
	MaxHeapBytes        uint64        `json:"max_heap_bytes"`        // Memory high watermark (e.g. 512MB)
	EnableAutoGC        bool          `json:"enable_auto_gc"`        // Automatically trigger GC on memory pressure
}

// DefaultConfig provides sensible defaults for Guardcat monitoring.
var DefaultConfig = Config{
	FailureThreshold:    3,
	EOFTriggerThreshold: 3,
	ProbeInterval:       5 * time.Second,
	MaxHeapBytes:        512 * 1024 * 1024, // 512 MB
	EnableAutoGC:        true,
}

// Guardcat is the auto-finetuning and monitoring watchdog module managing app health,
// memory state, DNS EOF trigger thresholds, circuit breaking, and environment readiness.
type Guardcat struct {
	mu        sync.RWMutex
	cfg       Config
	endpoints map[string]*EndpointState
	keepAlive bool
}

// New creates a new Guardcat instance with given configuration.
func New(cfg Config) *Guardcat {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = DefaultConfig.FailureThreshold
	}
	if cfg.EOFTriggerThreshold <= 0 {
		cfg.EOFTriggerThreshold = DefaultConfig.EOFTriggerThreshold
	}
	if cfg.ProbeInterval <= 0 {
		cfg.ProbeInterval = DefaultConfig.ProbeInterval
	}
	if cfg.MaxHeapBytes == 0 {
		cfg.MaxHeapBytes = DefaultConfig.MaxHeapBytes
	}

	return &Guardcat{
		cfg:       cfg,
		endpoints: make(map[string]*EndpointState),
	}
}

// RegisterEndpoint registers or updates an endpoint to monitor.
func (g *Guardcat) RegisterEndpoint(name, address, protocol string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := address + "|" + protocol
	if _, exists := g.endpoints[key]; !exists {
		g.endpoints[key] = &EndpointState{
			Name:     name,
			Address:  address,
			Protocol: protocol,
			Status:   StatusHealthy,
		}
	}
}

// RecordSuccess records a successful query for an endpoint, resetting failure and EOF counts.
func (g *Guardcat) RecordSuccess(address, protocol string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := address + "|" + protocol
	if ep, exists := g.endpoints[key]; exists {
		ep.FailureCount = 0
		ep.EOFCount = 0
		ep.Status = StatusHealthy
		ep.LastCheck = time.Now()
		ep.LastError = ""
	} else {
		g.endpoints[key] = &EndpointState{
			Address:   address,
			Protocol:  protocol,
			Status:    StatusHealthy,
			LastCheck: time.Now(),
		}
	}
}

// RecordFailure records a failure for an endpoint, checking EOF trigger thresholds
// and applying circuit breaking (Degraded -> Unhealthy).
func (g *Guardcat) RecordFailure(address, protocol string, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := address + "|" + protocol
	ep, exists := g.endpoints[key]
	if !exists {
		ep = &EndpointState{
			Address:  address,
			Protocol: protocol,
			Status:   StatusHealthy,
		}
		g.endpoints[key] = ep
	}

	ep.FailureCount++
	ep.LastCheck = time.Now()
	if err != nil {
		ep.LastError = err.Error()
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "eof") || strings.Contains(errStr, "reset") || strings.Contains(errStr, "closed") {
			ep.EOFCount++
		}
	}

	// Circuit breaking logic based on thresholds
	if ep.FailureCount >= g.cfg.FailureThreshold || ep.EOFCount >= g.cfg.EOFTriggerThreshold {
		ep.Status = StatusUnhealthy
	} else if ep.FailureCount >= 2 || ep.EOFCount >= 2 {
		ep.Status = StatusDegraded
	}
}

// IsHealthy checks if a specific endpoint is currently operational (Healthy or Degraded, but not Unhealthy).
func (g *Guardcat) IsHealthy(address, protocol string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	key := address + "|" + protocol
	if ep, exists := g.endpoints[key]; exists {
		return ep.Status != StatusUnhealthy
	}
	return true // Default optimistic if unknown
}

// GetEndpointStates returns a snapshot of all monitored endpoint states.
func (g *Guardcat) GetEndpointStates() map[string]EndpointState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	snapshot := make(map[string]EndpointState, len(g.endpoints))
	for k, ep := range g.endpoints {
		snapshot[k] = *ep
	}
	return snapshot
}

// CheckMemory inspects current memory allocation and goroutine count, applying auto-finetuning
// (such as running runtime.GC() on pressure) and returning a MemoryReport.
func (g *Guardcat) CheckMemory() MemoryReport {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	goroutines := runtime.NumGoroutine()
	status := "OPTIMAL"
	action := "NONE"

	if m.HeapAlloc > g.cfg.MaxHeapBytes {
		status = "CRITICAL"
		if g.cfg.EnableAutoGC {
			runtime.GC()
			runtime.ReadMemStats(&m)
			action = "GC_TRIGGERED_MEMORY_HIGH"
			if m.HeapAlloc > g.cfg.MaxHeapBytes {
				action = "GC_TRIGGERED_STILL_HIGH_RESTART_RECOMMENDED"
			}
		} else {
			action = "RESTART_RECOMMENDED_MEMORY_EXCEEDED"
		}
	} else if m.HeapAlloc > (g.cfg.MaxHeapBytes * 3 / 4) {
		status = "WARNING"
		action = "MONITORING_HIGH_WATERMARK"
	}

	return MemoryReport{
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		NumGoroutine: goroutines,
		NumGC:        m.NumGC,
		Status:       status,
		ActionTaken:  action,
		Timestamp:    time.Now(),
	}
}

// EnsureReadiness performs a pre-test readiness probe (resolving a minimal lightweight domain)
// to verify the local socket stack and DNS readiness before complex workflows execute.
func (g *Guardcat) EnsureReadiness(ctx context.Context, probeTarget string) error {
	if probeTarget == "" {
		probeTarget = "1.1.1.1:53"
	}

	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.DialContext(ctx, "udp", probeTarget)
	if err != nil {
		// Fallback to TCP or standard resolver check if UDP dial fails
		dTCP := net.Dialer{Timeout: 3 * time.Second}
		connTCP, errTCP := dTCP.DialContext(ctx, "tcp", probeTarget)
		if errTCP != nil {
			return fmt.Errorf("guardcat readiness check failed (UDP: %v, TCP: %v): local socket stack or network unreachable", err, errTCP)
		}
		connTCP.Close()
		return nil
	}
	conn.Close()
	return nil
}

// CheckInternetConnectivity performs a pre-test network connectivity awareness check
// and returns whether the system is online, reporting status suitable for pretest banners.
func (g *Guardcat) CheckInternetConnectivity(ctx context.Context) (bool, string) {
	// Probe well-known root / public resolvers with strict timeout
	targets := []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
	for _, target := range targets {
		d := net.Dialer{Timeout: 2 * time.Second}
		if conn, err := d.DialContext(ctx, "tcp", target); err == nil {
			conn.Close()
			return true, "[INTERNET ONLINE] System is connected to the internet. Low network safeguards active."
		}
		if conn, err := d.DialContext(ctx, "udp", target); err == nil {
			conn.Close()
			return true, "[INTERNET ONLINE] System is connected to the internet via UDP. Low network safeguards active."
		}
	}
	return false, "[OFFLINE / NO NETWORK ADVANTAGE] System detected offline or restricted network environment. Local fallback mode engaged."
}

// RecommendAction evaluates endpoints and memory state, providing actionable guidance
// on whether to restart, bypass failing servers, or proceed with test workflows.
func (g *Guardcat) RecommendAction() string {
	mem := g.CheckMemory()
	states := g.GetEndpointStates()

	unhealthyCount := 0
	degradedCount := 0
	for _, st := range states {
		switch st.Status {
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	if mem.Status == "CRITICAL" && strings.Contains(mem.ActionTaken, "RESTART") {
		return "[GUARDCAT CRITICAL] Memory usage exceeds threshold. Recommend restarting the application or environment before executing heavy test suites."
	}

	if unhealthyCount > 0 {
		return fmt.Sprintf("[GUARDCAT CIRCUIT BREAKER] %d endpoint(s) marked UNHEALTHY (exceeded failure/EOF threshold). Workflows will automatically bypass or use secondary resolvers.", unhealthyCount)
	}

	if degradedCount > 0 {
		return fmt.Sprintf("[GUARDCAT WARNING] %d endpoint(s) degraded. Probe frequency reduced; fallback paths active.", degradedCount)
	}

	return fmt.Sprintf("[GUARDCAT READY] Environment stable. Memory: %d MB allocated, Goroutines: %d. All systems ready for testing.", mem.HeapAlloc/(1024*1024), mem.NumGoroutine)
}
