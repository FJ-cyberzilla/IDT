package config

import (
	"testing"
)

func TestDefaultOptionsAndDialers(t *testing.T) {
	opts := DefaultOptions()
	if opts.Target == "" {
		t.Error("expected default target to be set")
	}
	if opts.Dials <= 0 {
		t.Error("expected positive default dials")
	}
	if opts.Timeout <= 0 {
		t.Error("expected positive timeout")
	}
	if opts.MaxDials <= 0 {
		t.Error("expected positive max dials")
	}

	dialers := DefaultDialers()
	if len(dialers) == 0 {
		t.Error("expected default dialers list to be non-empty")
	}

	// Verify all 3 protocols (Standard, DoT, DoH) are represented
	protoCount := map[string]int{}
	for _, d := range dialers {
		protoCount[d.Protocol]++
		if d.Name == "" || d.Address == "" {
			t.Errorf("dialer missing Name or Address: %+v", d)
		}
	}
	if protoCount[ProtoStandard] == 0 || protoCount[ProtoDoT] == 0 || protoCount[ProtoDoH] == 0 {
		t.Error("expected dialers across Standard, DoT, and DoH protocols")
	}
}

func TestSafeTargets(t *testing.T) {
	if len(SafeTargets) == 0 {
		t.Error("expected SafeTargets to be non-empty")
	}
	for _, target := range SafeTargets {
		if target == "" {
			t.Error("found empty safe target")
		}
	}
}

func TestPolicies(t *testing.T) {
	if len(Policies) == 0 {
		t.Error("expected Policies map to be non-empty")
	}
	for name, p := range Policies {
		if p.Name == "" {
			t.Errorf("policy %s missing Name", name)
		}
		if p.PrivacyScore < 1 || p.PrivacyScore > 5 {
			t.Errorf("policy %s has invalid PrivacyScore %d", name, p.PrivacyScore)
		}
		if p.AndroidDoT == "" {
			t.Errorf("policy %s missing AndroidDoT", name)
		}
	}
}

func TestFilterByPreference(t *testing.T) {
	dialers := DefaultDialers()

	t.Run("privacy", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "privacy")
		if len(filtered) == 0 {
			t.Error("expected non-empty dialers for 'privacy'")
		}
		for _, d := range filtered {
			p := Policies[d.Name]
			if p.PrivacyScore < 4 {
				t.Errorf("dialer %s has privacy score %d, expected >= 4", d.Name, p.PrivacyScore)
			}
		}
	})

	t.Run("no-logging", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "no-logging")
		if len(filtered) == 0 {
			t.Error("expected non-empty dialers for 'no-logging'")
		}
		for _, d := range filtered {
			p := Policies[d.Name]
			if p.LogDays != 0 {
				t.Errorf("dialer %s has LogDays %d, expected 0", d.Name, p.LogDays)
			}
		}
	})

	t.Run("blocking", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "blocking")
		for _, d := range filtered {
			p := Policies[d.Name]
			if p.Blocking == "none" {
				t.Errorf("dialer %s has blocking 'none', expected filtering", d.Name)
			}
		}
	})

	t.Run("unfiltered", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "unfiltered")
		for _, d := range filtered {
			p := Policies[d.Name]
			if p.Blocking != "none" {
				t.Errorf("dialer %s has blocking %s, expected 'none'", d.Name, p.Blocking)
			}
		}
	})

	t.Run("swiss", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "swiss")
		for _, d := range filtered {
			p := Policies[d.Name]
			if p.Jurisdiction != "CH" {
				t.Errorf("dialer %s has jurisdiction %s, expected 'CH'", d.Name, p.Jurisdiction)
			}
		}
	})

	t.Run("canadian", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "canadian")
		for _, d := range filtered {
			p := Policies[d.Name]
			if p.Jurisdiction != "CA" {
				t.Errorf("dialer %s has jurisdiction %s, expected 'CA'", d.Name, p.Jurisdiction)
			}
		}
	})

	t.Run("empty preference", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "")
		if len(filtered) != len(dialers) {
			t.Error("expected all dialers when preference is empty")
		}
	})

	t.Run("unknown preference", func(t *testing.T) {
		filtered := FilterByPreference(dialers, "unknown-pref")
		// Falls back or returns empty allowed -> returns all dialers in implementation
		if len(filtered) == 0 {
			t.Error("expected fallback behavior for unknown preference")
		}
	})
}
