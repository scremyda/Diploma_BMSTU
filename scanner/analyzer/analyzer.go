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
		return fmt.Errorf("certificate for %s expires in %s (alert threshold: %s)",
			scrape.Target, scrape.ExpiresIn, a.conf.AlertInterval)
	}

	parsedURL, err := url.Parse(scrape.Target)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %v", scrape.Target, err)
	}
	host := parsedURL.Scheme

	expectedPattern := a.conf.OverrideCN
	if expectedPattern == "" {
		expectedPattern = scrape.CN
	}

	if !matchesDomain(expectedPattern, host) && !contains(scrape.SANs, host) {
		return fmt.Errorf("certificate for %s has unexpected CN: got %s, expected %s", scrape.Target, scrape.CN, host)
	}

	return nil
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

func contains(sans []string, host string) bool {
	for _, s := range sans {
		if matchesDomain(s, host) {
			return true
		}
	}
	return false
}
