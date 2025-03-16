package scraper

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"time"
)

type Conf struct {
	Target        string        `yaml:"target"`
	Interval      time.Duration `yaml:"interval"`
	AlertInterval time.Duration `yaml:"alert_interval"`
	OverrideCN    string        `yaml:"override_cn"`
}

type Scraper struct {
}

type Result struct {
	Target    string
	ExpiresIn time.Duration
	CN        string
	Errors    error
}

func NewScraper() *Scraper {
	return &Scraper{}
}

func (r *Scraper) Scrape(ctx context.Context, targetURL string) Result {
	result := Result{Target: targetURL}
	u, err := url.Parse(targetURL)
	if err != nil {
		result.Errors = fmt.Errorf("invalid URL: %w", err)

		return result
	}

	host := u.Hostname()
	if host == "" {
		result.Errors = fmt.Errorf("empty host in target")

		return result
	}

	port := u.Port()
	if port == "" {
		port = "443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: 10 * time.Second,
		},
		"tcp",
		host+":"+port,
		&tls.Config{
			InsecureSkipVerify: true,
		},
	)
	if err != nil {
		result.Errors = fmt.Errorf("connection failed: %w", err)

		return result
	}
	defer conn.Close()

	if len(conn.ConnectionState().PeerCertificates) == 0 {
		result.Errors = fmt.Errorf("no peer certificates found")

		return result
	}

	cert := conn.ConnectionState().PeerCertificates[0]
	result.CN = cert.Subject.CommonName
	result.ExpiresIn = time.Until(cert.NotAfter)

	return result
}
