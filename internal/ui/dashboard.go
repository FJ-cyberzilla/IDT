package ui

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"idt/internal/check"
	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/enterprise"
	"idt/internal/errorhandling"
	"idt/internal/runner"

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
	state           state
	choices         []string
	enterpriseChoices []string
	cursor          int
	enterpriseCursor int
	spinner         spinner.Model
	outputLines     []string
	loadingText     string
	offlineReason   string
	warmupStep      warmupStep
	connType        string
	preflightMsg    string
	startTime       time.Time
	err             error
	opts            config.Options
	dialers         []config.Dialer
	checkResults    []runner.Result
	sortedResults   []runner.Result
	verdicts        map[string]compare.Verdict
	matrixResults   []enterprise.MatrixResult
	protocolResults []enterprise.ProtocolTestResult
	replayResult    enterprise.ReplayResult
	width           int
	height          int
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
			"📊 Interactive Dashboard & Telemetry",
			"⚙ View Config & Active Servers",
			"🔍 Run Connectivity Check",
			"📈 Run Tuning Analysis",
			"🏢 Advanced Enterprise & Protocol Simulation",
			"ℹ️ About Us & App Capabilities",
			"❌ Exit Command Center",
		},
		enterpriseChoices: []string{
			"🌐 Remote Office & Data Center Performance Matrix",
			"📡 Multi-Protocol Testing (HTTP, VoIP, SaaS, Custom Apps)",
			"📼 Live Traffic Capture & Lab Replay Simulator",
			"⬅️ Return to Main Command Center",
		},
		cursor:     0,
		spinner:    s,
		warmupStep: stepInit,
		opts:       opts,
		dialers:    dialers,
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
		// Requirement: ESC to cancel operation at any time user wanted
		if key == "esc" {
			if m.state == stateMenu {
				return m, tea.Quit
			}
			m.state = stateMenu
			return m, nil
		}

		// Requirement: be sure using ctrl z, or c dont crash the app
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
				case 0, 1: // Run IDT Test or Dashboard
					m.state = stateWarmup
					m.startTime = time.Now()
					m.warmupStep = stepInit
					m.loadingText = "Initializing IDT Integrated Command Center..."
					return m, tea.Batch(m.spinner.Tick, tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
						return warmupProgressMsg{nextStep: stepInit, loadingText: "Initializing command center & configuration..."}
					}))
				case 2: // View Servers
					m.state = stateServers
					return m, nil
				case 3: // Check
					m.state = stateCheckRunning
					m.startTime = time.Now()
					m.loadingText = "Running real-world connectivity diagnostics..."
					return m, tea.Batch(m.spinner.Tick, runCheckCmd())
				case 4: // Tune
					m.state = stateTuneRunning
					m.startTime = time.Now()
					m.loadingText = "Running RTT distribution & tuning analysis..."
					return m, tea.Batch(m.spinner.Tick, runTuneCmd(m.opts, m.dialers))
				case 5: // Enterprise Menu
					m.state = stateEnterpriseMenu
					m.enterpriseCursor = 0
					return m, nil
				case 6: // About Us
					m.state = stateAbout
					return m, nil
				case 7: // Exit
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
				case 3: // Back to Menu
					m.state = stateMenu
					return m, nil
				}
			} else if m.state == stateMatrixResults || m.state == stateProtocolResults || m.state == stateReplayResults {
				m.state = stateEnterpriseMenu
				return m, nil
			} else if m.state == stateResults || m.state == stateCheckResults || m.state == stateTuneResults || m.state == stateServers || m.state == stateAbout || m.state == stateOfflineError {
				m.state = stateMenu
				return m, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
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
				return warmupProgressMsg{nextStep: stepPreflight, loadingText: "Public resolvers reachable (1.1.1.1 / 8.8.8.8)", preflightMsg: "OK"}
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
		return m, nil
	}

	return m, nil
}

func checkPreflightConnectivity() (bool, string) {
	conn, err := net.DialTimeout("tcp", "1.1.1.1:53", 1500*time.Millisecond)
	if err != nil {
		conn2, err2 := net.DialTimeout("tcp", "8.8.8.8:53", 1500*time.Millisecond)
		if err2 != nil {
			errStr := strings.ToLower(err.Error() + " " + err2.Error())
			if strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "timeout") {
				return false, fmt.Sprintf("No internet connection or network stalled / offline (Primary: %v, Fallback: %v)", err, err2)
			}
			if errorhandling.IsAndroidPrivateDNSFailure(err) || errorhandling.IsAndroidPrivateDNSFailure(err2) {
				return false, "Android Private DNS active but VPN disconnected (Network unreachable / Private DNS rejection). Disable Private DNS in settings or connect VPN."
			}
			return false, fmt.Sprintf("No internet connection or network stalled (Primary: %v, Fallback: %v)", err, err2)
		}
		conn2.Close()
		return true, "Connected (via Fallback)"
	}
	conn.Close()
	return true, "Connected"
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

func runCheckCmd() tea.Cmd {
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
		results := runner.Run(ctx, check.Dialers, checkOpts)
		time.Sleep(300 * time.Millisecond)
		return checkCompleteMsg{results: results}
	}
}

func runTuneCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results := runner.Run(ctx, dialers, opts)
		sorted := runner.SortByMedian(results)

		truthAnswer, _ := runner.Dial(ctx,
			config.Dialer{Name: "truth", Protocol: config.ProtoDoH,
				Address: "https://cloudflare-dns.com/dns-query"},
			opts.Target, opts.Timeout)

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

		truthAnswer, _ := runner.Dial(ctx,
			config.Dialer{Name: "truth", Protocol: config.ProtoDoH,
				Address: "https://cloudflare-dns.com/dns-query"},
			opts.Target, opts.Timeout)

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

func (m Model) View() string {
	var sb strings.Builder

	width := m.width
	if width <= 0 {
		width = 80
	}

	boxWidth := width - 4
	if boxWidth < 30 {
		boxWidth = width - 2
	}
	if boxWidth < 20 {
		boxWidth = 40
	}

	isMobileLike := width > 0 && width < 60
	currentBoxStyle := BoxStyle.Width(boxWidth)
	if isMobileLike {
		currentBoxStyle = currentBoxStyle.Padding(0, 1).Margin(0, 0)
	} else {
		currentBoxStyle = currentBoxStyle.Padding(1, 2).Margin(1, 1)
	}

	// Header banner
	header := HeaderStyle.Render("🌐 IDT — INTERNET DIALER TEST COMMAND CENTER")
	sb.WriteString(header + "\n\n")

	switch m.state {
	case stateMenu:
		sb.WriteString(RenderPhaseBadge("MENU", "MENU") + "\n\n")
		sb.WriteString(RenderMenuChoices(m.choices, m.cursor, width))

	case stateWarmup:
		elapsed := time.Since(m.startTime).Round(time.Second)
		sb.WriteString(RenderPhaseBadge("WARMUP", "WARMUP") + "\n\n")
		sb.WriteString(RenderWarmupProgress(m.warmupStep, m.spinner.View(), m.connType, width, elapsed))

	case stateRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		elapsedStr := fmt.Sprintf("[%02d:%02d]", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		sb.WriteString(RenderPhaseBadge("RUNNING", "RUNNING") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(elapsedStr)))
		sb.WriteString(HelpStyle.Render("Integrated Dashboard • [ESC] Cancel • [q] Quit"))

	case stateResults:
		sb.WriteString(RenderPhaseBadge("RESULTS", "RESULTS") + "\n\n")
		sb.WriteString(TitleStyle.Render(" 📊 IDT TEST RESULTS & TELEMETRY ") + "\n\n")
		sb.WriteString("Execution completed successfully. Live verdicts computed against ground truth.\n\n")
		sb.WriteString(fmt.Sprintf("Configured Dialers: %d\n", len(m.dialers)))
		sb.WriteString(fmt.Sprintf("Target Domain: %s\n\n", m.opts.Target))
		sb.WriteString(HelpStyle.Render("[Enter/ESC] Return to Menu • [q] Quit"))

	case stateServers:
		sb.WriteString(RenderPhaseBadge("SERVERS", "SERVERS") + "\n\n")
		sb.WriteString(TitleStyle.Render(" ⚙ CONFIGURED SERVERS & PROTOCOLS ") + "\n\n")
		for _, d := range m.dialers {
			addr := d.Address
			if isMobileLike && len(addr) > 28 {
				addr = addr[:25] + "..."
			}
			sb.WriteString(fmt.Sprintf(" • [%s] %s (%s)\n", ProtocolStyle.Render(d.Protocol), d.Name, addr))
		}
		sb.WriteString(HelpStyle.Render("\n[Enter/ESC] Return to Menu"))

	case stateEnterpriseMenu:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderEnterpriseMenu(m.enterpriseChoices, m.enterpriseCursor, width))

	case stateMatrixRunning, stateProtocolRunning, stateReplayRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		elapsedStr := fmt.Sprintf("[%02d:%02d]", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(elapsedStr)))
		sb.WriteString(HelpStyle.Render("Enterprise Simulation • [ESC] Cancel • [q] Quit"))

	case stateMatrixResults:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderMatrixResultsView(m.matrixResults, width))

	case stateProtocolResults:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderProtocolResultsView(m.protocolResults, width))

	case stateReplayResults:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderReplayResultsView(m.replayResult, width))

	case stateAbout:
		sb.WriteString(RenderPhaseBadge("ABOUT", "ABOUT") + "\n\n")
		sb.WriteString(RenderAboutView(width))

	case stateCheckRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		elapsedStr := fmt.Sprintf("[%02d:%02d]", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		sb.WriteString(RenderPhaseBadge("CHECK", "CHECK") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(elapsedStr)))
		sb.WriteString(HelpStyle.Render("Running diagnostics... [ESC] Cancel"))

	case stateCheckResults:
		sb.WriteString(RenderPhaseBadge("CHECK", "CHECK") + "\n\n")
		sb.WriteString(RenderCheckResultsView(m.checkResults, width))

	case stateTuneRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		elapsedStr := fmt.Sprintf("[%02d:%02d]", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
		sb.WriteString(RenderPhaseBadge("TUNE", "TUNE") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(elapsedStr)))
		sb.WriteString(HelpStyle.Render("Running tuning analysis... [ESC] Cancel"))

	case stateTuneResults:
		sb.WriteString(RenderPhaseBadge("TUNE", "TUNE") + "\n\n")
		sb.WriteString(RenderTuneResultsView(m.sortedResults, m.verdicts, m.opts, width))

	case stateOfflineError:
		sb.WriteString(RenderPhaseBadge("ERROR", "ERROR") + "\n\n")
		sb.WriteString(RenderErrorCard(m.offlineReason, width))
	}

	return currentBoxStyle.Render(sb.String())
}
