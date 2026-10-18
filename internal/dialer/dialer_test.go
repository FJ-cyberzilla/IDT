package dialer

import (
	"net"
	"testing"
	"time"
)

func TestParseA_ShortPacket(t *testing.T) {
	_, err := ParseA([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for short packet, got nil")
	}
}

func TestParseA_NoAnswers(t *testing.T) {
	buf := []byte{
		0x12, 0x34, // ID
		0x01, 0x00, // Flags
		0x00, 0x01, // QDCOUNT
		0x00, 0x00, // ANCOUNT = 0
		0x00, 0x00, // NSCOUNT
		0x00, 0x00, // ARCOUNT
		7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		3, 'c', 'o', 'm', 0,
		0x00, 0x01, // TYPE A
		0x00, 0x01, // CLASS IN
	}
	_, err := ParseA(buf)
	if err == nil {
		t.Error("expected error for zero answers, got nil")
	}
}

func TestParseA_ValidRecord(t *testing.T) {
	buf := []byte{
		0x12, 0x34, // ID
		0x81, 0x80, // Flags (Response, RA)
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x01, // ANCOUNT = 1
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
		// Question
		7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		3, 'c', 'o', 'm', 0,
		0x00, 0x01, // TYPE A
		0x00, 0x01, // CLASS IN
		// Answer
		0xc0, 0x0c, // Name pointer to question
		0x00, 0x01, // TYPE A
		0x00, 0x01, // CLASS IN
		0x00, 0x00, 0x00, 0x3c, // TTL = 60
		0x00, 0x04, // RDLENGTH = 4
		192, 0, 2, 1, // RDATA = 192.0.2.1
	}

	ips, err := ParseA(buf)
	if err != nil {
		t.Fatalf("unexpected error parsing valid A record: %v", err)
	}
	if len(ips) != 1 {
		t.Fatalf("expected 1 IP, got %d", len(ips))
	}
	if !ips[0].Equal(net.IPv4(192, 0, 2, 1)) {
		t.Errorf("expected 192.0.2.1, got %s", ips[0])
	}
}

func TestBuildQuery(t *testing.T) {
	query := buildQuery("example.com")
	if len(query) < 12 {
		t.Error("query packet too short")
	}
}

func TestStandardDialerCreation(t *testing.T) {
	d := Standard{Address: "8.8.8.8:53", Timeout: 2 * time.Second}
	if d.Address != "8.8.8.8:53" {
		t.Errorf("expected address 8.8.8.8:53, got %s", d.Address)
	}
}

func TestDoHDialerCreation(t *testing.T) {
	d := DoH{URL: "https://cloudflare-dns.com/dns-query", Timeout: 2 * time.Second}
	if d.URL != "https://cloudflare-dns.com/dns-query" {
		t.Errorf("expected URL, got %s", d.URL)
	}
}

func TestDoTDialerCreation(t *testing.T) {
	d := DoT{Address: "one.one.one.one:853", Timeout: 2 * time.Second}
	if d.Address != "one.one.one.one:853" {
		t.Errorf("expected address, got %s", d.Address)
	}
}

func TestDialerInterfaceCompliance(t *testing.T) {
	var _ Dialer = Standard{}
	var _ Dialer = DoH{}
	var _ Dialer = DoT{}
}
