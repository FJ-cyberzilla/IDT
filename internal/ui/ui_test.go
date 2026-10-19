package ui

import (
	"testing"
	"idt/internal/enterprise"
)

func TestInitialModel(t *testing.T) {
	model := InitialModel()
	if view := model.View(); view == "" {
		t.Error("expected non-empty view string from model")
	}
}

func TestRenderVisualComponents(t *testing.T) {
	width := 80

	// Test Phase Badges
	if badge := RenderPhaseBadge("MENU", "MENU"); badge == "" {
		t.Error("expected non-empty menu badge")
	}
	if badge := RenderPhaseBadge("ABOUT", "ABOUT"); badge == "" {
		t.Error("expected non-empty about badge")
	}

	// Test About View
	aboutStr := RenderAboutView(width)
	if aboutStr == "" {
		t.Error("expected non-empty about view")
	}

	// Test Enterprise Menu
	choices := []string{"Choice 1", "Choice 2"}
	menuStr := RenderEnterpriseMenu(choices, 0, width)
	if menuStr == "" {
		t.Error("expected non-empty enterprise menu")
	}

	// Test Matrix Results
	matrixRes := []enterprise.MatrixResult{
		{
			Source:      enterprise.DefaultLocations[0],
			Destination: enterprise.DefaultLocations[1],
			Latency:     25,
			Loss:        0.0,
			Throughput:  1000.0,
			SLAStatus:   "OPTIMAL",
		},
	}
	matrixStr := RenderMatrixResultsView(matrixRes, width)
	if matrixStr == "" {
		t.Error("expected non-empty matrix results view")
	}

	// Test Protocol Results
	protoRes := []enterprise.ProtocolTestResult{
		{
			Config:      enterprise.DefaultProtocolTests[0],
			SuccessRate: 100.0,
			AvgLatency:  20,
			Score:       95,
			Status:      "HEALTHY",
			Details:     "OK",
		},
	}
	protoStr := RenderProtocolResultsView(protoRes, width)
	if protoStr == "" {
		t.Error("expected non-empty protocol results view")
	}

	// Test Replay Results
	replayRes := enterprise.ReplayResult{
		TotalCaptured:   5,
		TotalReplayed:   5,
		DroppedPackets:  0,
		AnomalyDetected: false,
		LogSummary:      []string{"test log"},
	}
	replayStr := RenderReplayResultsView(replayRes, width)
	if replayStr == "" {
		t.Error("expected non-empty replay results view")
	}
}
