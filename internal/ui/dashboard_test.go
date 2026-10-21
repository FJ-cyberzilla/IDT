package ui

import (
	"testing"

	"idt/internal/compare"
	"idt/internal/runner"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDashboardInitialModelAndInit(t *testing.T) {
	m := InitialModel()
	if m.state != stateMenu {
		t.Errorf("expected initial state stateMenu, got %v", m.state)
	}
	if len(m.choices) == 0 {
		t.Error("expected non-empty menu choices")
	}

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected non-nil tea.Cmd from Init()")
	}
}

func TestDashboardWindowResize(t *testing.T) {
	m := InitialModel()
	newModel, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if cmd != nil {
		t.Error("expected nil cmd on window resize")
	}
	um := newModel.(Model)
	if um.width != 100 || um.height != 40 {
		t.Errorf("expected dimensions 100x40, got %dx%d", um.width, um.height)
	}
}

func TestDashboardKeyNavigation(t *testing.T) {
	tests := []struct {
		name           string
		initialState   state
		initialCursor  int
		msg            tea.Msg
		expectedState  state
		expectedCursor int
		expectQuit     bool
	}{
		{
			name:           "Menu down navigation",
			initialState:   stateMenu,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedState:  stateMenu,
			expectedCursor: 1,
		},
		{
			name:           "Menu up navigation (at top)",
			initialState:   stateMenu,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedState:  stateMenu,
			expectedCursor: 0,
		},
		{
			name:           "Menu down arrow navigation",
			initialState:   stateMenu,
			initialCursor:  1,
			msg:            tea.KeyMsg{Type: tea.KeyDown},
			expectedState:  stateMenu,
			expectedCursor: 2,
		},
		{
			name:           "Menu up arrow navigation",
			initialState:   stateMenu,
			initialCursor:  2,
			msg:            tea.KeyMsg{Type: tea.KeyUp},
			expectedState:  stateMenu,
			expectedCursor: 1,
		},
		{
			name:           "Quit via q in menu",
			initialState:   stateMenu,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			expectedState:  stateMenu,
			expectedCursor: 0,
			expectQuit:     true,
		},
		{
			name:           "ESC in menu quits",
			initialState:   stateMenu,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyEsc},
			expectedState:  stateMenu,
			expectedCursor: 0,
			expectQuit:     true,
		},
		{
			name:           "ESC in non-menu returns to menu",
			initialState:   stateAbout,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyEsc},
			expectedState:  stateMenu,
			expectedCursor: 0,
		},
		{
			name:           "Ctrl+C in menu quits",
			initialState:   stateMenu,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyCtrlC},
			expectedState:  stateMenu,
			expectedCursor: 0,
			expectQuit:     true,
		},
		{
			name:           "Ctrl+Z is noop",
			initialState:   stateMenu,
			initialCursor:  0,
			msg:            tea.KeyMsg{Type: tea.KeyCtrlZ},
			expectedState:  stateMenu,
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := InitialModel()
			m.state = tt.initialState
			m.cursor = tt.initialCursor

			newModel, cmd := m.Update(tt.msg)
			um := newModel.(Model)

			if um.state != tt.expectedState {
				t.Errorf("expected state %v, got %v", tt.expectedState, um.state)
			}
			if um.cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, um.cursor)
			}
			if tt.expectQuit && cmd == nil {
				t.Error("expected quit command, got nil")
			}
		})
	}
}

func TestDashboardCustomMessages(t *testing.T) {
	m := InitialModel()

	// testCompleteMsg
	tModel, _ := m.Update(testCompleteMsg{
		sorted:   []runner.Result{{Total: 1}},
		verdicts: map[string]compare.Verdict{"Test": {Checked: true, Matches: 1, Total: 1}},
	})
	tm := tModel.(Model)
	if tm.state != stateResults {
		t.Errorf("expected stateResults, got %v", tm.state)
	}

	// checkCompleteMsg
	cModel, _ := m.Update(checkCompleteMsg{
		results: []runner.Result{{Total: 1}},
	})
	cm := cModel.(Model)
	if cm.state != stateCheckResults {
		t.Errorf("expected stateCheckResults, got %v", cm.state)
	}

	// tuneCompleteMsg
	tuModel, _ := m.Update(tuneCompleteMsg{
		sorted:   []runner.Result{{Total: 1}},
		verdicts: map[string]compare.Verdict{"Tune": {Checked: true, Matches: 1, Total: 1}},
	})
	tum := tuModel.(Model)
	if tum.state != stateTuneResults {
		t.Errorf("expected stateTuneResults, got %v", tum.state)
	}

	// warmupProgressMsg
	wModel, _ := m.Update(warmupProgressMsg{
		nextStep:     stepNetwork,
		loadingText:  "Loading...",
		connType:     "WiFi",
		preflightMsg: "OK",
		failed:       false,
	})
	wm := wModel.(Model)
	if wm.warmupStep != stepNetwork {
		t.Errorf("expected warmupStep stepNetwork, got %v", wm.warmupStep)
	}
}

func TestDashboardViewRendering(t *testing.T) {
	m := InitialModel()
	m.width = 80
	m.height = 30

	// Test menu view
	m.state = stateMenu
	if view := m.View(); view == "" {
		t.Error("expected non-empty menu view")
	}

	// Test about view
	m.state = stateAbout
	if view := m.View(); view == "" {
		t.Error("expected non-empty about view")
	}

	// Test offline error view
	m.state = stateOfflineError
	m.offlineReason = "Network unreachable"
	if view := m.View(); view == "" {
		t.Error("expected non-empty offline error view")
	}

	// Test enterprise menu view
	m.state = stateEnterpriseMenu
	if view := m.View(); view == "" {
		t.Error("expected non-empty enterprise menu view")
	}
}
