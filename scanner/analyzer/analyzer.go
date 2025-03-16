package analyzer

import (
	"context"
	"diploma/scanner/scraper"
	"fmt"
	"net/url"
)

type Analyzer struct {
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(ctx context.Context, conf scraper.Conf, result scraper.Result) []string {
	var warnings []string

	if result.Errors != nil {
		warnings = append(warnings, fmt.Sprintf("Error for %s: %v", result.Target, result.Errors))
		return warnings
	}

	if result.ExpiresIn < conf.AlertInterval {
		warnings = append(warnings, fmt.Sprintf("Certificate for %s expires in %s (alert threshold: %s)",
			result.Target, result.ExpiresIn, conf.AlertInterval))
	}

	parsedURL, err := url.Parse(conf.Target)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Invalid URL %s: %v", conf.Target, err))
		return warnings
	}

	expectedCN := result.CN
	if conf.OverrideCN != "" {
		expectedCN = conf.OverrideCN
	}

	expectedHost := parsedURL.Hostname()

	if !matchesDomain(expectedCN, expectedHost) {
		warnings = append(warnings, fmt.Sprintf("Certificate for %s has unexpected CN: got %s, expected %s",
			conf.Target, result.CN, expectedHost))
	}

	return warnings
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
