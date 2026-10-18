package ui

import (
	"testing"
)

func TestInitialModel(t *testing.T) {
	model := InitialModel()
	if view := model.View(); view == "" {
		t.Error("expected non-empty view string from model")
	}
}
