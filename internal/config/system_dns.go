package config

import (
	"bufio"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemDNSConfig holds discovered system, Android, Windows, WSL, Termux, DoH, and DoT DNS configurations.
type SystemDNSConfig struct {
	ResolvConfServers []string
	AndroidDNS        []string
	WindowsDNS        []string
	WSLDNS            []string
	AndroidPrivateDNS string
	DoHServers        []string
	DoTServers        []string
}

// DiscoverSystemDNS examines what is configured in Android, Windows (PowerShell/ipconfig), WSL 2, Termux (/etc/resolv.conf), DoH, and DoT.
func DiscoverSystemDNS() SystemDNSConfig {
	cfg := SystemDNSConfig{
		ResolvConfServers: []string{},
		AndroidDNS:        []string{},
		WindowsDNS:        []string{},
		WSLDNS:            []string{},
		DoHServers:        []string{},
		DoTServers:        []string{},
	}

	// 1. Check Termux / Linux / WSL /etc/resolv.conf
	if file, err := os.Open("/etc/resolv.conf"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "nameserver") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					ip := fields[1]
					if net.ParseIP(ip) != nil {
						cfg.ResolvConfServers = append(cfg.ResolvConfServers, ip)
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
				} else if strings.Contains(line, ".") && !strings.Contains(line, "Description") && !strings.Contains(line, "Adapter") {
					val := strings.TrimSpace(line)
					if net.ParseIP(val) != nil {
						cfg.WindowsDNS = append(cfg.WindowsDNS, val)
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
		} else if out, err := exec.Command("powershell.exe", "-Command", "(Get-DnsClientServerAddress -AddressFamily IPv4).ServerAddresses").Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				ip := strings.TrimSpace(scanner.Text())
				if net.ParseIP(ip) != nil {
					cfg.WSLDNS = append(cfg.WSLDNS, ip)
				}
			}
		}
	}

	// 3. Check Android system properties via getprop (if available)
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
		if out, err := exec.Command("getprop", "net.private.dns.specifier").Output(); err == nil {
			spec := strings.TrimSpace(string(out))
			if spec != "" {
				cfg.AndroidPrivateDNS = spec
				cfg.DoTServers = append(cfg.DoTServers, spec+":853")
			}
		}
	}

	// 4. Environment or default DoH / DoT hints
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
