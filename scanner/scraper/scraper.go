package scraper

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type Conf struct {
	Target  string        `yaml:"target"`
	Timeout time.Duration `yaml:"timeout"`
}

type Scraper struct {
	conf Conf
}

type Scrape struct {
	Target    string
	ExpiresIn time.Duration
	CN        string
	SANs      []string
}

func NewScraper(cfg Conf) *Scraper {
	return &Scraper{
		conf: cfg,
	}
}

func (r *Scraper) Scrape(ctx context.Context) (Scrape, error) {
	u, err := url.Parse(r.conf.Target)
	if err != nil {
		return Scrape{}, fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Hostname()
	if len(host) == 0 {
		return Scrape{}, errors.New("empty host in target")
	}

	var addr strings.Builder
	addr.WriteString(host)

	port := u.Port()
	if len(port) != 0 {
		addr.WriteString(":")
		addr.WriteString(port)
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: r.conf.Timeout,
		},
		"tcp",
		addr.String(),
		&tls.Config{
			InsecureSkipVerify: true,
		},
	)
	if err != nil {
		return Scrape{}, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	if len(conn.ConnectionState().PeerCertificates) == 0 {
		return Scrape{}, errors.New("no peer certificates found")
	}
	cert := conn.ConnectionState().PeerCertificates[0]

	return Scrape{
		Target:    addr.String(),
		CN:        cert.Subject.CommonName,
		ExpiresIn: time.Until(cert.NotAfter),
		SANs:      cert.DNSNames,
	}, nil
}
