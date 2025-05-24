package analyzer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"diploma/scanner/scraper"
)

type Conf struct {
	AlertInterval time.Duration `yaml:"alert_interval"`
	OverrideCN    string        `yaml:"override_cn"`
}

type Analyzer struct {
	conf Conf
}

func NewAnalyzer(cfg Conf) *Analyzer {
	return &Analyzer{conf: cfg}
}

func (a *Analyzer) Analyze(ctx context.Context, scrape scraper.Scrape) error {
	if scrape.ExpiresIn < a.conf.AlertInterval {
		reason := fmt.Sprintf("сертификат истекает через %s", scrape.ExpiresIn)
		msg := formatAnalysisError(scrape, "Просрочка", reason)
		return errors.New(msg)
	}

	parsedURL, err := url.Parse(scrape.Target)
	if err != nil {
		reason := "invalid URL"
		detail := err.Error()
		msg := formatAnalysisError(scrape, reason, detail)
		return errors.New(msg)
	}
	host := parsedURL.Scheme

	expected := a.conf.OverrideCN
	if expected == "" {
		expected = scrape.CN
	}
	if !matchesDomain(expected, host) && !contains(scrape.SANs, host) {
		reason := "неверный CN/SAN"
		detail := fmt.Sprintf("получили %q, ожидаем %q", scrape.CN, host)
		msg := formatAnalysisError(scrape, reason, detail)
		return errors.New(msg)
	}

	return nil
}

func formatAnalysisError(scrape scraper.Scrape, reason, detail string) string {
	return fmt.Sprintf(
		"⚠️ Проблема с сертификатом! ⚠️\n\n"+
			"• Домен: `%s`\n"+
			"• Причина: %s\n"+
			"• Детали: %s\n",
		scrape.Target,
		reason,
		detail,
	)
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
