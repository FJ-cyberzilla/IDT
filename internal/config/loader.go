package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"
)

type fileFormat struct {
	Options struct {
		Target      string `json:"target"`
		Dials       int    `json:"dials"`
		TimeoutMS   int    `json:"timeout_ms"`
		JitterMinMS int    `json:"jitter_min_ms"`
		JitterMaxMS int    `json:"jitter_max_ms"`
		MaxDials    int    `json:"max_dials"`
	} `json:"options"`
	Dialers []Dialer `json:"dialers"`
}

func Load() (Options, []Dialer, error) {
	opts := DefaultOptions()
	dialers := DefaultDialers()

	var (
		configPath = flag.String("config", "servers.json", "path to dialer JSON")
		target     = flag.String("target", "", "domain to dial (overrides JSON)")
		dials      = flag.Int("dials", 0, "dials per endpoint (overrides JSON)")
		timeoutMS  = flag.Int("timeout", 0, "timeout in ms (overrides JSON)")
		plan       = flag.Bool("plan", false, "show plan without sending packets")
		safety     = flag.Bool("safety", false, "strict safe defaults regardless of config")
		check      = flag.Bool("check", false, "run minimal safe health check (5 dials)")
		tune       = flag.Bool("tune", false, "print tuning recommendation for this network")
		prefer     = flag.String("dial", "", "privacy|no-logging|blocking|unfiltered|swiss|canadian")
		tag        = flag.String("tag", "", "tag this run")
	)
	flag.Parse()

	if data, err := os.ReadFile(*configPath); err == nil {
		var f fileFormat
		if err := json.Unmarshal(data, &f); err == nil {
			if f.Options.Target != "" {
				opts.Target = f.Options.Target
			}
			if f.Options.Dials > 0 {
				opts.Dials = f.Options.Dials
			}
			if f.Options.TimeoutMS > 0 {
				opts.Timeout = time.Duration(f.Options.TimeoutMS) * time.Millisecond
			}
			if f.Options.JitterMinMS > 0 {
				opts.JitterMin = time.Duration(f.Options.JitterMinMS) * time.Millisecond
			}
			if f.Options.JitterMaxMS > 0 {
				opts.JitterMax = time.Duration(f.Options.JitterMaxMS) * time.Millisecond
			}
			if f.Options.MaxDials > 0 {
				opts.MaxDials = f.Options.MaxDials
			}
			if len(f.Dialers) > 0 {
				dialers = f.Dialers
			}
		}
	}

	if *target != "" {
		opts.Target = *target
	}
	if *dials > 0 {
		opts.Dials = *dials
	}
	if *timeoutMS > 0 {
		opts.Timeout = time.Duration(*timeoutMS) * time.Millisecond
	}

	opts.Plan = *plan
	opts.Safety = *safety
	opts.Check = *check
	opts.Tune = *tune
	opts.Prefer = *prefer
	opts.Tag = *tag

	if opts.Safety {
		opts.Target = "example.com"
		if opts.Dials > 3 {
			opts.Dials = 3
		}
		opts.JitterMin = 300 * time.Millisecond
		opts.JitterMax = 900 * time.Millisecond
		opts.MaxDials = 90
	}

	return opts, dialers, nil
}
