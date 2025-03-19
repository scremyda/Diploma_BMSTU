package analyzer

import (
	"context"
	"diploma/scanner/scraper"
	"fmt"
	"net/url"
	"time"
)

type Conf struct {
	AlertInterval time.Duration `yaml:"alert_interval"`
	OverrideCN    string        `yaml:"override_cn"`
}

type Analyzer struct {
	conf Conf
}

func NewAnalyzer(cfg Conf) *Analyzer {
	return &Analyzer{
		conf: cfg,
	}
}

func (a *Analyzer) Analyze(ctx context.Context, scrape scraper.Scrape) error {
	if scrape.ExpiresIn < a.conf.AlertInterval {
		return fmt.Errorf("Certificate for %s expires in %s (alert threshold: %s)",
			scrape.Target, scrape.ExpiresIn, a.conf.AlertInterval)

	}

	parsedURL, err := url.Parse(scrape.Target)
	if err != nil {
		return fmt.Errorf("Invalid URL %s: %v", scrape.Target, err)
	}

	expectedCN := scrape.CN
	if a.conf.OverrideCN != "" {
		expectedCN = a.conf.OverrideCN
	}

	expectedHost := parsedURL.Scheme
	if !matchesDomain(expectedCN, expectedHost) {
		err = fmt.Errorf("certificate for %s has unexpected CN: got %s, expected %s",
			scrape.Target, scrape.CN, expectedHost)
	}

	return err
}

func matchesDomain(pattern, host string) bool {
	if pattern == host {
		return true
	}

	if len(pattern) > 2 && pattern[0] == '*' && pattern[1] == '.' {
		domain := pattern[2:]
		return len(host) >= len(domain) && host[len(host)-len(domain):] == domain
	}

	return false
}
