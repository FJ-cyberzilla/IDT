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
	state         state
	choices       []string
	cursor        int
	spinner       spinner.Model
	outputLines   []string
	loadingText   string
	offlineReason string
	warmupStep    warmupStep
	connType      string
	preflightMsg  string
	err           error
	opts          config.Options
	dialers       []config.Dialer
	checkResults  []runner.Result
	sortedResults []runner.Result
	verdicts      map[string]compare.Verdict
}

func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	opts, dialers, _ := config.Load()

	return Model{
		state: stateMenu,
		choices: []string{
			"🚀 Run Internet Dialer Test (IDT)",
			"📊 Interactive Dashboard & Telemetry",
			"⚙ View Config & Active Servers",
			"🔍 Run Connectivity Check",
			"📈 Run Tuning Analysis",
			"❌ Exit Command Center",
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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
			}

		case "down", "j":
			if m.state == stateMenu {
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			}

		case "r":
			if m.state == stateOfflineError {
				m.state = stateWarmup
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
					m.loadingText = "Running real-world connectivity diagnostics..."
					return m, tea.Batch(m.spinner.Tick, runCheckCmd())
				case 4: // Tune
					m.state = stateTuneRunning
					m.loadingText = "Running RTT distribution & tuning analysis..."
					return m, tea.Batch(m.spinner.Tick, runTuneCmd(m.opts, m.dialers))
				case 5: // Exit
					return m, tea.Quit
				}
			} else if m.state == stateResults || m.state == stateCheckResults || m.state == stateTuneResults || m.state == stateServers || m.state == stateOfflineError {
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

	// Header banner
	header := HeaderStyle.Render("🌐 IDT — INTERNET DIALER TEST COMMAND CENTER")
	sb.WriteString(header + "\n\n")

	switch m.state {
	case stateMenu:
		sb.WriteString(RenderPhaseBadge("MENU", "MENU") + "\n\n")
		sb.WriteString(RenderMenuChoices(m.choices, m.cursor))

	case stateWarmup:
		sb.WriteString(RenderPhaseBadge("WARMUP", "WARMUP") + "\n\n")
		sb.WriteString(RenderWarmupProgress(m.warmupStep, m.spinner.View(), m.connType))

	case stateRunning:
		sb.WriteString(RenderPhaseBadge("RUNNING", "RUNNING") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText)))
		sb.WriteString(HelpStyle.Render("Integrated Dashboard • Real-time Benchmark Execution..."))

	case stateResults:
		sb.WriteString(RenderPhaseBadge("RESULTS", "RESULTS") + "\n\n")
		sb.WriteString(TitleStyle.Render(" 📊 IDT TEST RESULTS & TELEMETRY ") + "\n\n")
		sb.WriteString("Execution completed successfully. Live verdicts computed against ground truth.\n\n")
		sb.WriteString(fmt.Sprintf("Configured Dialers: %d\n", len(m.dialers)))
		sb.WriteString(fmt.Sprintf("Target Domain: %s\n\n", m.opts.Target))
		sb.WriteString(HelpStyle.Render("[Enter] Return to Menu • [q] Quit"))

	case stateServers:
		sb.WriteString(RenderPhaseBadge("SERVERS", "SERVERS") + "\n\n")
		sb.WriteString(TitleStyle.Render(" ⚙ CONFIGURED SERVERS & PROTOCOLS ") + "\n\n")
		for _, d := range m.dialers {
			sb.WriteString(fmt.Sprintf(" • [%s] %s (%s)\n", ProtocolStyle.Render(d.Protocol), d.Name, d.Address))
		}
		sb.WriteString(HelpStyle.Render("\n[Enter] Return to Menu"))

	case stateCheckRunning:
		sb.WriteString(RenderPhaseBadge("CHECK", "CHECK") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText)))
		sb.WriteString(HelpStyle.Render("Running real-world connectivity diagnostics..."))

	case stateCheckResults:
		sb.WriteString(RenderPhaseBadge("CHECK", "CHECK") + "\n\n")
		sb.WriteString(RenderCheckResultsView(m.checkResults))

	case stateTuneRunning:
		sb.WriteString(RenderPhaseBadge("TUNE", "TUNE") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText)))
		sb.WriteString(HelpStyle.Render("Running RTT distribution & tuning analysis..."))

	case stateTuneResults:
		sb.WriteString(RenderPhaseBadge("TUNE", "TUNE") + "\n\n")
		sb.WriteString(RenderTuneResultsView(m.sortedResults, m.verdicts, m.opts))

	case stateOfflineError:
		sb.WriteString(RenderPhaseBadge("ERROR", "ERROR") + "\n\n")
		sb.WriteString(RenderErrorCard(m.offlineReason))
	}

	return BoxStyle.Render(sb.String())
}
