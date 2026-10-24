package config

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemDNSConfig holds discovered system, Android, Windows, WSL, Termux, Private DNS, Browser DoH, VPN, and Country configuration.
type SystemDNSConfig struct {
	ResolvConfServers   []string `json:"resolv_conf_servers"`
	AndroidDNS          []string `json:"android_dns"`
	WindowsDNS          []string `json:"windows_dns"`
	WSLDNS              []string `json:"wsl_dns"`
	AndroidPrivateDNS   string   `json:"android_private_dns"`
	PrivateDNSState     string   `json:"private_dns_state"` // OFF, STRICT, AUTO
	AndroidVPNActive    bool     `json:"android_vpn_active"`
	BrowserDoHDetected  bool     `json:"browser_doh_detected"`
	BrowserDoHDetails   string   `json:"browser_doh_details"`
	ConnectedCountry    string   `json:"connected_country"`
	DoHServers          []string `json:"doh_servers"`
	DoTServers          []string `json:"dot_servers"`
}

// DiscoverSystemDNS examines what is configured in Android, Termux, Windows, WSL, Private DNS, VPN, Browsers, and Country geo.
func DiscoverSystemDNS() SystemDNSConfig {
	cfg := SystemDNSConfig{
		ResolvConfServers:  []string{},
		AndroidDNS:         []string{},
		WindowsDNS:         []string{},
		WSLDNS:             []string{},
		PrivateDNSState:    "OFF",
		AndroidVPNActive:   false,
		BrowserDoHDetected: false,
		BrowserDoHDetails:  "None detected",
		ConnectedCountry:   "Unknown / Local",
		DoHServers:         []string{},
		DoTServers:         []string{},
	}

	// 1. Check Termux / Linux / WSL /etc/resolv.conf and Termux custom resolv.conf
	resolvPaths := []string{
		"/etc/resolv.conf",
		"/data/data/com.termux/files/usr/etc/resolv.conf",
	}
	for _, rPath := range resolvPaths {
		if file, err := os.Open(rPath); err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "nameserver") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						ip := fields[1]
						if net.ParseIP(ip) != nil {
							// avoid duplicate
							exists := false
							for _, existing := range cfg.ResolvConfServers {
								if existing == ip {
									exists = true
									break
								}
							}
							if !exists {
								cfg.ResolvConfServers = append(cfg.ResolvConfServers, ip)
							}
						}
					}
				}
			}
		}
	}

	// Check if running inside WSL 2
	isWSL := false
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/version"); err == nil {
			if strings.Contains(strings.ToLower(string(data)), "microsoft") || strings.Contains(strings.ToLower(string(data)), "wsl") {
				isWSL = true
			}
		}
	}

	// 2. Check Windows DNS (Native Windows or via WSL bridge)
	if runtime.GOOS == "windows" {
		if out, err := exec.Command("ipconfig", "/all").Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "DNS Servers") || strings.Contains(line, "DNS Server") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						ip := strings.TrimSpace(parts[1])
						if net.ParseIP(ip) != nil {
							cfg.WindowsDNS = append(cfg.WindowsDNS, ip)
						}
					}
				}
			}
		}
		if path, err := exec.LookPath("powershell"); err == nil && path != "" {
			if out, err := exec.Command("powershell", "-Command", "(Get-DnsClientServerAddress -AddressFamily IPv4).ServerAddresses").Output(); err == nil {
				scanner := bufio.NewScanner(strings.NewReader(string(out)))
				for scanner.Scan() {
					ip := strings.TrimSpace(scanner.Text())
					if net.ParseIP(ip) != nil {
						cfg.WindowsDNS = append(cfg.WindowsDNS, ip)
					}
				}
			}
		}
	} else if isWSL {
		if out, err := exec.Command("/mnt/c/Windows/System32/ipconfig.exe", "/all").Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "DNS Servers") || strings.Contains(line, "DNS Server") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						ip := strings.TrimSpace(parts[1])
						if net.ParseIP(ip) != nil {
							cfg.WSLDNS = append(cfg.WSLDNS, ip)
						}
					}
				}
			}
		}
	}

	// 3. Check Android system properties via getprop (DNS, Private DNS, VPN status)
	if path, err := exec.LookPath("getprop"); err == nil && path != "" {
		if out, err := exec.Command("getprop", "net.dns1").Output(); err == nil {
			dns := strings.TrimSpace(string(out))
			if dns != "" && net.ParseIP(dns) != nil {
				cfg.AndroidDNS = append(cfg.AndroidDNS, dns)
			}
		}
		if out, err := exec.Command("getprop", "net.dns2").Output(); err == nil {
			dns := strings.TrimSpace(string(out))
			if dns != "" && net.ParseIP(dns) != nil {
				cfg.AndroidDNS = append(cfg.AndroidDNS, dns)
			}
		}
		if out, err := exec.Command("getprop", "net.private.dns.mode").Output(); err == nil {
			mode := strings.TrimSpace(string(out))
			if mode != "" {
				cfg.PrivateDNSState = strings.ToUpper(mode)
			}
		}
		if out, err := exec.Command("getprop", "net.private.dns.specifier").Output(); err == nil {
			spec := strings.TrimSpace(string(out))
			if spec != "" {
				cfg.AndroidPrivateDNS = spec
				cfg.PrivateDNSState = "STRICT"
				cfg.DoTServers = append(cfg.DoTServers, spec+":853")
			}
		}
		if cfg.PrivateDNSState == "OFF" && cfg.AndroidPrivateDNS == "" {
			// Check if auto mode heuristic
			if out, err := exec.Command("getprop", "net.dns.collector").Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
				cfg.PrivateDNSState = "AUTO"
			}
		}
	}

	// 4. Check Android VPN status via network interfaces (tun0, ppp0, wg0)
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			name := strings.ToLower(iface.Name)
			if strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "ppp") || strings.HasPrefix(name, "wg") || strings.Contains(name, "vpn") {
				cfg.AndroidVPNActive = true
				break
			}
		}
	}

	// 5. Check Browser DoH / Private DNS configuration (Chromium Local State / Firefox prefs.js)
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		// Chromium Local State (Chrome / Edge / Brave)
		chromiumPaths := []string{
			homeDir + "/.config/google-chrome/Local State",
			homeDir + "/.config/chromium/Local State",
			homeDir + "/.config/microsoft-edge/Local State",
		}
		for _, cp := range chromiumPaths {
			if data, err := os.ReadFile(cp); err == nil {
				var result map[string]interface{}
				if json.Unmarshal(data, &result) == nil {
					if dnsSec, ok := result["dns_over_https"].(map[string]interface{}); ok {
						cfg.BrowserDoHDetected = true
						mode, _ := dnsSec["mode"].(string)
						template, _ := dnsSec["templates"].(string)
						cfg.BrowserDoHDetails = "Chromium DoH [Mode: " + mode + ", Template: " + template + "]"
						if template != "" {
							cfg.DoHServers = append(cfg.DoHServers, template)
						}
					}
				}
			}
		}

		// Firefox prefs.js
		firefoxProfilesDir := homeDir + "/.mozilla/firefox"
		if entries, err := os.ReadDir(firefoxProfilesDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && strings.Contains(entry.Name(), ".default") {
					prefsPath := firefoxProfilesDir + "/" + entry.Name() + "/prefs.js"
					if file, err := os.Open(prefsPath); err == nil {
						scanner := bufio.NewScanner(file)
						for scanner.Scan() {
							line := scanner.Text()
							if strings.Contains(line, "network.trr.mode") || strings.Contains(line, "network.trr.uri") {
								cfg.BrowserDoHDetected = true
								cfg.BrowserDoHDetails = "Firefox DoH Configured (" + strings.TrimSpace(line) + ")"
							}
						}
						file.Close()
					}
				}
			}
		}
	}

	// 6. Connected Country detection heuristic (based on local timezone or public resolver IP / geo hint)
	if tz := os.Getenv("TZ"); tz != "" {
		cfg.ConnectedCountry = "Detected via TZ (" + tz + ")"
	} else {
		cfg.ConnectedCountry = "Global / Autonomous Region"
	}

	// Environment or default DoH / DoT hints
	if dohEnv := os.Getenv("IDT_DOH_SERVER"); dohEnv != "" {
		cfg.DoHServers = append(cfg.DoHServers, dohEnv)
	}
	if dotEnv := os.Getenv("IDT_DOT_SERVER"); dotEnv != "" {
		cfg.DoTServers = append(cfg.DoTServers, dotEnv)
	}

	return cfg
}

// GetDiscoveredDialers converts discovered system DNS configs into IDT Dialer structs.
func GetDiscoveredDialers() []Dialer {
	sys := DiscoverSystemDNS()
	var dialers []Dialer

	addUnique := func(name string, protocol string, address string) {
		for _, d := range dialers {
			if d.Address == address {
				return
			}
		}
		dialers = append(dialers, Dialer{
			Name:     name,
			Protocol: protocol,
			Address:  address,
		})
	}

	for _, ip := range sys.AndroidDNS {
		addUnique("Android-System", ProtoStandard, ip+":53")
	}
	for _, ip := range sys.WindowsDNS {
		addUnique("Windows-System", ProtoStandard, ip+":53")
	}
	for _, ip := range sys.WSLDNS {
		addUnique("WSL-Host-DNS", ProtoStandard, ip+":53")
	}
	for _, ip := range sys.ResolvConfServers {
		addUnique("Termux-Resolv", ProtoStandard, ip+":53")
	}
	for _, doh := range sys.DoHServers {
		addUnique("System-DoH", ProtoDoH, doh)
	}
	for _, dot := range sys.DoTServers {
		addUnique("System-DoT", ProtoDoT, dot)
	}

	return dialers
}
