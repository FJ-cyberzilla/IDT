package ui

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/enterprise"
	"idt/internal/errorhandling"
	"idt/internal/runner"
	"idt/internal/timediag"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateMenu state = iota
	stateWarmup
	stateRunning
	stateResults
	stateDashboard
	stateCheckRunning
	stateCheckResults
	stateTuneRunning
	stateTuneResults
	stateServers
	stateEnterpriseMenu
	stateMatrixRunning
	stateMatrixResults
	stateProtocolRunning
	stateProtocolResults
	stateReplayRunning
	stateReplayResults
	stateColdStartRunning
	stateColdStartResults
	stateHotlineRunning
	stateHotlineResults
	stateDNSSECRunning
	stateDNSSECResults
	stateTimezoneRunning
	stateTimezoneResults
	stateCorruptionRunning
	stateCorruptionResults
	stateTimeDiagRunning
	stateTimeDiagResults
	stateAbout
	stateOfflineError
)

type warmupStep int

const (
	stepInit warmupStep = iota
	stepNetwork
	stepPreflight
	stepReady
	stepDone
)

type Model struct {
	state             state
	choices           []string
	enterpriseChoices []string
	cursor            int
	enterpriseCursor  int
	spinner           spinner.Model
	outputLines       []string
	loadingText       string
	offlineReason     string
	warmupStep        warmupStep
	connType          string
	preflightMsg      string
	startTime         time.Time
	err               error
	opts              config.Options
	dialers           []config.Dialer
	checkResults      []runner.Result
	sortedResults     []runner.Result
	verdicts          map[string]compare.Verdict
	matrixResults     []enterprise.MatrixResult
	protocolResults   []enterprise.ProtocolTestResult
	replayResult      enterprise.ReplayResult
	coldStartResults  []runner.Result
	hotlineReport     enterprise.ProtocolTestResult
	dnssecResult      enterprise.DNSSECResult
	timezoneReports   []enterprise.TimezoneReport
	corruptionResult  enterprise.CorruptionResult
	timeDiagReport    timediag.WorkflowReport
	guardcatPhase     string
	guardcatDotCount  int
	width             int
	height            int
}

func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	opts := config.DefaultOptions()
	dialers := config.NormalizeDialers(config.DefaultDialers())

	return Model{
		state: stateMenu,
		choices: []string{
			"🚀 Run Internet Dialer Test (IDT)",
			"❄️ Cold Start Test (Current Config)",
			"📊 Interactive Dashboard & Telemetry",
			"⚙ View Config & Active Servers",
			"🔍 Run Connectivity Check",
			"📈 Run Tuning Analysis",
			"🏢 Advanced Enterprise & Protocol Simulation",
			"⏳ Time-Diag Network Workflow Engine (Chronological)",
			"ℹ️ About Us & App Capabilities",
			"❌ Exit Command Center",
		},
		enterpriseChoices: []string{
			"🌐 Remote Office & Data Center Performance Matrix",
			"📡 Multi-Protocol Testing (HTTP, VoIP, SaaS, Custom Apps)",
			"📼 Live Traffic Capture & Lab Replay Simulator",
			"🛡️ DNSSEC Chain-of-Trust Audit",
			"🔍 DNS Corruption Detection & Auto-Reconnect Audit",
			"🌍 Timezone-Aware Regional Routing Analysis",
			"🔥 Hotline Test (High Congestion Times under Load)",
			"⬅️ Return to Main Command Center",
		},
		cursor:           0,
		spinner:          s,
		warmupStep:       stepInit,
		opts:             opts,
		dialers:          dialers,
		guardcatPhase:    "preparing",
		guardcatDotCount: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

type testCompleteMsg struct {
	sorted   []runner.Result
	verdicts map[string]compare.Verdict
}

type checkCompleteMsg struct {
	results []runner.Result
}

type tuneCompleteMsg struct {
	sorted   []runner.Result
	verdicts map[string]compare.Verdict
}

type warmupProgressMsg struct {
	nextStep     warmupStep
	loadingText  string
	connType     string
	preflightMsg string
	failed       bool
	failReason   string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		if key == "esc" {
			if m.state == stateMenu {
				return m, tea.Quit
			}
			m.state = stateMenu
			return m, nil
		}

		if key == "ctrl+c" {
			if m.state == stateMenu || m.state == stateOfflineError {
				return m, tea.Quit
			}
			m.state = stateMenu
			return m, nil
		}

		if key == "ctrl+z" {
			return m, nil
		}

		switch key {
		case "q":
			if m.state == stateMenu || m.state == stateOfflineError {
				return m, tea.Quit
			}
			m.state = stateMenu
			return m, nil

		case "up", "k":
			if m.state == stateMenu {
				if m.cursor > 0 {
					m.cursor--
				}
			} else if m.state == stateEnterpriseMenu {
				if m.enterpriseCursor > 0 {
					m.enterpriseCursor--
				}
			}

		case "down", "j":
			if m.state == stateMenu {
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			} else if m.state == stateEnterpriseMenu {
				if m.enterpriseCursor < len(m.enterpriseChoices)-1 {
					m.enterpriseCursor++
				}
			}

		case "r":
			if m.state == stateOfflineError {
				m.state = stateWarmup
				m.startTime = time.Now()
				m.warmupStep = stepInit
				m.loadingText = "Retrying IDT Warmup & Network Verification..."
				return m, tea.Batch(m.spinner.Tick, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
					return warmupProgressMsg{nextStep: stepInit, loadingText: "Initializing command center & configuration..."}
				}))
			}

		case "enter":
			if m.state == stateMenu {
				switch m.cursor {
				case 0: // Run IDT Test
					m.state = stateWarmup
					m.startTime = time.Now()
					m.warmupStep = stepInit
					m.loadingText = "Initializing IDT Integrated Command Center..."
					return m, tea.Batch(m.spinner.Tick, tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
						return warmupProgressMsg{nextStep: stepInit, loadingText: "Initializing command center & configuration..."}
					}))
				case 1: // Cold Start Test
					m.state = stateColdStartRunning
					m.startTime = time.Now()
					m.loadingText = "Executing Cold Start Test with zero-cache sockets..."
					return m, tea.Batch(m.spinner.Tick, runColdStartCmd(m.opts, m.dialers))
				case 2: // Interactive Dashboard & Telemetry
					if len(m.sortedResults) == 0 {
						m.state = stateWarmup
						m.startTime = time.Now()
						m.warmupStep = stepInit
						m.loadingText = "Initializing Interactive Dashboard Telemetry..."
						return m, tea.Batch(m.spinner.Tick, tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
							return warmupProgressMsg{nextStep: stepInit, loadingText: "Initializing telemetry inspection..."}
						}))
					}
					m.state = stateDashboard
					return m, nil
				case 3: // View Servers
					m.state = stateServers
					return m, nil
				case 4: // Check
					m.state = stateCheckRunning
					m.startTime = time.Now()
					m.loadingText = "Running real-world connectivity diagnostics..."
					return m, tea.Batch(m.spinner.Tick, runCheckCmd(m.dialers))
				case 5: // Tune
					m.state = stateTuneRunning
					m.startTime = time.Now()
					m.loadingText = "Running RTT distribution & tuning analysis..."
					return m, tea.Batch(m.spinner.Tick, runTuneCmd(m.opts, m.dialers))
				case 6: // Enterprise Menu
					m.state = stateEnterpriseMenu
					m.enterpriseCursor = 0
					return m, nil
				case 7: // Time-Diag Workflow Engine
					m.state = stateTimeDiagRunning
					m.startTime = time.Now()
					m.loadingText = "Executing Time-Diag Chronological Network & DNS Workflow Engine..."
					return m, tea.Batch(m.spinner.Tick, runTimeDiagCmd(m.opts, m.dialers))
				case 8: // About Us
					m.state = stateAbout
					return m, nil
				case 9: // Exit
					return m, tea.Quit
				}
			} else if m.state == stateEnterpriseMenu {
				switch m.enterpriseCursor {
				case 0: // Matrix
					m.state = stateMatrixRunning
					m.startTime = time.Now()
					m.loadingText = "Simulating Remote Office & Data Center Performance Matrix..."
					return m, tea.Batch(m.spinner.Tick, runMatrixCmd())
				case 1: // Protocols
					m.state = stateProtocolRunning
					m.startTime = time.Now()
					m.loadingText = "Running Multi-Protocol Synthetic Probes under Dynamic Conditions..."
					return m, tea.Batch(m.spinner.Tick, runProtocolCmd())
				case 2: // Replay
					m.state = stateReplayRunning
					m.startTime = time.Now()
					m.loadingText = "Capturing and Replaying Live Production Traffic in Lab Setting..."
					return m, tea.Batch(m.spinner.Tick, runReplayCmd())
				case 3: // DNSSEC
					m.state = stateDNSSECRunning
					m.startTime = time.Now()
					m.loadingText = "Auditing DNSSEC Chain-of-Trust and Signature Validation..."
					return m, tea.Batch(m.spinner.Tick, runDNSSECCmd())
				case 4: // Corruption Detection
					m.state = stateCorruptionRunning
					m.startTime = time.Now()
					m.loadingText = "Auditing DNS Response Integrity & Auto-reconnect Triggers..."
					return m, tea.Batch(m.spinner.Tick, runCorruptionCmd())
				case 5: // Timezone
					m.state = stateTimezoneRunning
					m.startTime = time.Now()
					m.loadingText = "Analyzing Timezone-Aware Regional Routing Latencies..."
					return m, tea.Batch(m.spinner.Tick, runTimezoneCmd())
				case 6: // Hotline
					m.state = stateHotlineRunning
					m.startTime = time.Now()
					m.loadingText = "Running Hotline Test for High Congestion Times under Load..."
					return m, tea.Batch(m.spinner.Tick, runHotlineCmd())
				case 7: // Back to Menu
					m.state = stateMenu
					return m, nil
				}
			} else if m.state == stateMatrixResults || m.state == stateProtocolResults || m.state == stateReplayResults || m.state == stateDNSSECResults || m.state == stateTimezoneResults || m.state == stateHotlineResults || m.state == stateCorruptionResults {
				m.state = stateEnterpriseMenu
				return m, nil
			} else if m.state == stateResults || m.state == stateDashboard || m.state == stateCheckResults || m.state == stateTuneResults || m.state == stateServers || m.state == stateAbout || m.state == stateOfflineError || m.state == stateColdStartResults || m.state == stateTimeDiagResults {
				m.state = stateMenu
				return m, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.guardcatDotCount++
		if m.guardcatPhase == "preparing" && m.guardcatDotCount > 6 {
			m.guardcatPhase = "ready"
		}
		if m.state == stateOfflineError {
			m.guardcatPhase = "error"
		}
		return m, cmd

	case checkCompleteMsg:
		m.state = stateCheckResults
		m.checkResults = msg.results
		return m, nil

	case tuneCompleteMsg:
		m.state = stateTuneResults
		m.sortedResults = msg.sorted
		m.verdicts = msg.verdicts
		return m, nil

	case matrixCompleteMsg:
		m.state = stateMatrixResults
		m.matrixResults = msg.results
		return m, nil

	case protocolCompleteMsg:
		m.state = stateProtocolResults
		m.protocolResults = msg.results
		return m, nil

	case replayCompleteMsg:
		m.state = stateReplayResults
		m.replayResult = msg.result
		return m, nil

	case coldStartCompleteMsg:
		m.state = stateColdStartResults
		m.coldStartResults = msg.results
		return m, nil

	case hotlineCompleteMsg:
		m.state = stateHotlineResults
		m.hotlineReport = msg.report
		return m, nil

	case dnssecCompleteMsg:
		m.state = stateDNSSECResults
		m.dnssecResult = msg.result
		return m, nil

	case timezoneCompleteMsg:
		m.state = stateTimezoneResults
		m.timezoneReports = msg.reports
		return m, nil

	case corruptionCompleteMsg:
		m.state = stateCorruptionResults
		m.corruptionResult = msg.result
		return m, nil

	case timeDiagCompleteMsg:
		m.state = stateTimeDiagResults
		m.timeDiagReport = msg.report
		return m, nil

	case warmupProgressMsg:
		m.warmupStep = msg.nextStep
		m.loadingText = msg.loadingText
		if msg.connType != "" {
			m.connType = msg.connType
		}
		if msg.preflightMsg != "" {
			m.preflightMsg = msg.preflightMsg
		}

		if msg.failed {
			m.state = stateOfflineError
			m.offlineReason = msg.failReason
			return m, nil
		}

		switch m.warmupStep {
		case stepInit:
			return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				ct := detectConnectionType()
				return warmupProgressMsg{nextStep: stepNetwork, loadingText: fmt.Sprintf("Network detected: [%s]", ct), connType: ct}
			})
		case stepNetwork:
			return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				connected, reason := checkPreflightConnectivity()
				if !connected {
					return warmupProgressMsg{failed: true, failReason: reason}
				}
				return warmupProgressMsg{nextStep: stepPreflight, loadingText: "Public resolvers reachable", preflightMsg: "OK"}
			})
		case stepPreflight:
			return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				return warmupProgressMsg{nextStep: stepReady, loadingText: "Secure resolvers & connection pools prepared"}
			})
		case stepReady:
			return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				m.state = stateRunning
				m.startTime = time.Now()
				m.loadingText = "Executing DNS & IP Dialers across protocols..."
				return runTestCmd(m.opts, m.dialers)
			})
		}
		return m, nil

	case testCompleteMsg:
		m.state = stateResults
		m.sortedResults = msg.sorted
		m.verdicts = msg.verdicts
		return m, nil
	}

	return m, nil
}

func formatTimerWithETA(elapsed time.Duration, estimatedTotal time.Duration) string {
	remaining := estimatedTotal - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return fmt.Sprintf("[%02d:%02d / ETA %02d:%02d]",
		int(elapsed.Minutes()), int(elapsed.Seconds())%60,
		int(remaining.Minutes()), int(remaining.Seconds())%60)
}

func checkPreflightConnectivity() (bool, string) {
	sys := config.DiscoverSystemDNS()
	var testAddrs []string
	for _, ip := range sys.AndroidDNS {
		testAddrs = append(testAddrs, ip+":53")
	}
	for _, ip := range sys.ResolvConfServers {
		testAddrs = append(testAddrs, ip+":53")
	}
	for _, dot := range sys.DoTServers {
		testAddrs = append(testAddrs, dot)
	}
	testAddrs = append(testAddrs, "9.9.9.9:53", "185.228.168.168:53")

	var lastErr error
	for _, addr := range testAddrs {
		conn, err := net.DialTimeout("tcp", addr, 1500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true, fmt.Sprintf("Connected (via %s)", addr)
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no address resolved")
	}
	errStr := strings.ToLower(lastErr.Error())
	if strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "timeout") {
		return false, fmt.Sprintf("No internet connection or network stalled / offline (Last error: %v)", lastErr)
	}
	if errorhandling.IsAndroidPrivateDNSFailure(lastErr) {
		return false, "Android Private DNS active but VPN disconnected. Disable Private DNS or connect VPN."
	}
	return false, fmt.Sprintf("No internet connection or network stalled (Last error: %v)", lastErr)
}

func detectConnectionType() string {
	if ifs, err := net.Interfaces(); err == nil {
		for _, f := range ifs {
			name := strings.ToLower(f.Name)
			if strings.Contains(name, "tun") || strings.Contains(name, "tap") || strings.Contains(name, "wg") || strings.Contains(name, "ppp") || strings.Contains(name, "vpn") {
				return "VPN Active (" + f.Name + ")"
			}
		}
	}
	return "Normal ISP Connection"
}

func getTruthDialer(dialers []config.Dialer) config.Dialer {
	for _, d := range dialers {
		if d.Protocol == config.ProtoDoH {
			return d
		}
	}
	return config.Dialer{Name: "Quad9", Protocol: config.ProtoDoH, Address: "https://dns.quad9.net/dns-query"}
}

func runMatrixCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := enterprise.RunPerformanceMatrix(ctx, nil, 5)
		time.Sleep(300 * time.Millisecond)
		return matrixCompleteMsg{results: results}
	}
}

func runProtocolCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := enterprise.RunProtocolTests(ctx, nil)
		time.Sleep(300 * time.Millisecond)
		return protocolCompleteMsg{results: results}
	}
}

func runReplayCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		cfg := enterprise.ReplayConfig{
			SpeedMultiplier: 1.5,
			InjectJitter:    true,
			SimulateDrop:    true,
			BatchSize:       5,
		}
		result := enterprise.RunTrafficReplay(ctx, nil, cfg)
		time.Sleep(300 * time.Millisecond)
		return replayCompleteMsg{result: result}
	}
}

type matrixCompleteMsg struct {
	results []enterprise.MatrixResult
}

type protocolCompleteMsg struct {
	results []enterprise.ProtocolTestResult
}

type replayCompleteMsg struct {
	result enterprise.ReplayResult
}

type coldStartCompleteMsg struct {
	results []runner.Result
}

type hotlineCompleteMsg struct {
	report enterprise.ProtocolTestResult
}

type dnssecCompleteMsg struct {
	result enterprise.DNSSECResult
}

type timezoneCompleteMsg struct {
	reports []enterprise.TimezoneReport
}

type corruptionCompleteMsg struct {
	result enterprise.CorruptionResult
}

func runCheckCmd(dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		checkOpts := config.Options{
			Target:    "example.com",
			Dials:     1,
			Timeout:   3 * time.Second,
			JitterMin: 200 * time.Millisecond,
			JitterMax: 500 * time.Millisecond,
			MaxDials:  10,
		}
		results := runner.Run(ctx, dialers, checkOpts)
		time.Sleep(300 * time.Millisecond)
		return checkCompleteMsg{results: results}
	}
}

func runTuneCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := runner.Run(ctx, dialers, opts)
		sorted := runner.SortByMedian(results)

		truthDialer := getTruthDialer(dialers)
		truthAnswer, _ := runner.Dial(ctx, truthDialer, opts.Target, opts.Timeout)

		truthIPs := truthAnswer.IPs
		verdicts := map[string]compare.Verdict{}
		for _, r := range results {
			observed := mostCommonIPs(r.IPs)
			verdicts[r.Dialer.Name+"|"+r.Dialer.Protocol] = compare.Compare(observed, truthIPs)
		}

		time.Sleep(300 * time.Millisecond)
		return tuneCompleteMsg{sorted: sorted, verdicts: verdicts}
	}
}

func runTestCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := runner.Run(ctx, dialers, opts)
		sorted := runner.SortByMedian(results)

		truthDialer := getTruthDialer(dialers)
		truthAnswer, _ := runner.Dial(ctx, truthDialer, opts.Target, opts.Timeout)

		truthIPs := truthAnswer.IPs
		verdicts := map[string]compare.Verdict{}
		for _, r := range results {
			observed := mostCommonIPs(r.IPs)
			verdicts[r.Dialer.Name+"|"+r.Dialer.Protocol] = compare.Compare(observed, truthIPs)
		}

		time.Sleep(500 * time.Millisecond)
		return testCompleteMsg{sorted: sorted, verdicts: verdicts}
	}
}

func mostCommonIPs(samples [][]string) []net.IP {
	counts := map[string]int{}
	raw := map[string][]net.IP{}
	for _, ips := range samples {
		key := strings.Join(ips, ",")
		counts[key]++
		if _, ok := raw[key]; !ok {
			parsed := make([]net.IP, 0, len(ips))
			for _, s := range ips {
				if ip := net.ParseIP(s); ip != nil {
					parsed = append(parsed, ip)
				}
			}
			raw[key] = parsed
		}
	}
	var best string
	bestN := -1
	for k, n := range counts {
		if n > bestN {
			bestN, best = n, k
		}
	}
	return raw[best]
}



func runColdStartCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res := runner.Run(ctx, dialers, opts)
		return coldStartCompleteMsg{results: res}
	}
}

func runHotlineCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		cfg := enterprise.ProtocolTestConfig{
			Protocol:   enterprise.ProtoVoIP,
			TargetName: "Enterprise SIP Media Relay Hotline",
			Condition:  enterprise.ConditionCongested,
			Iterations: 15,
		}
		reps := enterprise.RunProtocolTests(ctx, []enterprise.ProtocolTestConfig{cfg})
		var rep enterprise.ProtocolTestResult
		if len(reps) > 0 {
			rep = reps[0]
		}
		return hotlineCompleteMsg{report: rep}
	}
}

func runDNSSECCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res := enterprise.AuditDNSSEC(ctx, "cloudflare-dns.com")
		return dnssecCompleteMsg{result: res}
	}
}

func runTimezoneCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		reps := enterprise.AnalyzeTimezones(ctx)
		return timezoneCompleteMsg{reports: reps}
	}
}

func runCorruptionCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res := enterprise.RunDNSCorruptionDetection(ctx, enterprise.DefaultCorruptionConfig)
		return corruptionCompleteMsg{result: res}
	}
}

type timeDiagCompleteMsg struct {
	report timediag.WorkflowReport
}

func runTimeDiagCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		engine := timediag.NewNetworkWorkflowEngine(opts, dialers)
		report, _ := engine.Run(ctx)
		return timeDiagCompleteMsg{report: report}
	}
}
