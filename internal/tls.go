/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"net/url"
	"time"
)

// TLSInfo holds TLS certificate expiry information
type TLSInfo struct {
	Expiry   time.Time
	DaysLeft int
}

// CheckTLSCertificate connects to the given URL and returns TLS certificate expiry info
func CheckTLSCertificate(rawURL string) (*TLSInfo, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("not an HTTPS URL")
	}

	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 5 * time.Second},
		"tcp",
		host+":"+port,
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	expiry := certs[0].NotAfter
	daysLeft := int(math.Ceil(time.Until(expiry).Hours() / 24))

	return &TLSInfo{
		Expiry:   expiry,
		DaysLeft: daysLeft,
	}, nil
}
