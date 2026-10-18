package check

import (
	"testing"
)

func TestDialersList(t *testing.T) {
	if len(Dialers) != 5 {
		t.Errorf("expected exactly 5 health-check dialers, got %d", len(Dialers))
	}
	for _, d := range Dialers {
		if d.Name == "" || d.Protocol == "" || d.Address == "" {
			t.Errorf("dialer missing fields: %+v", d)
		}
	}
}
