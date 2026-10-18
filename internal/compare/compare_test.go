package compare

import (
	"net"
	"testing"
)

func TestCompareMatch(t *testing.T) {
	observed := []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.2")}
	truth := []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.2")}

	v := Compare(observed, truth)
	if !v.Checked {
		t.Errorf("expected checked true")
	}
	if v.Matches != 2 || v.Total != 2 {
		t.Errorf("expected matches 2, total 2; got matches %d, total %d", v.Matches, v.Total)
	}
	if v.Label() != "match" {
		t.Errorf("expected label 'match', got %s", v.Label())
	}
}

func TestComparePartialAndMismatch(t *testing.T) {
	observed := []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.99")}
	truth := []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.2")}

	v := Compare(observed, truth)
	if v.Label() != "partial" {
		t.Errorf("expected label 'partial', got %s", v.Label())
	}

	observedMismatch := []net.IP{net.ParseIP("192.0.2.98"), net.ParseIP("192.0.2.99")}
	v2 := Compare(observedMismatch, truth)
	if v2.Label() != "mismatch" {
		t.Errorf("expected label 'mismatch', got %s", v2.Label())
	}
}
