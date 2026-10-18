package dialer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DoH struct {
	URL     string
	Timeout time.Duration
}

var dohClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (d DoH) Dial(ctx context.Context, target string) (Answer, error) {
	q := buildQuery(target)
	req, err := http.NewRequestWithContext(ctx, "POST", d.URL, bytes.NewReader(q))
	if err != nil {
		return Answer{}, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")
	req.Header.Set("User-Agent", "idt/1.0")

	start := time.Now()
	resp, err := dohClient.Do(req)
	if err != nil {
		return Answer{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	rtt := time.Since(start)

	if resp.StatusCode == http.StatusTooManyRequests {
		return Answer{}, fmt.Errorf("429 rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		return Answer{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	ips, _ := ParseA(body)
	return Answer{RTT: rtt, IPs: ips}, nil
}
