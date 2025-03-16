package scanner

import (
	"context"
	"diploma/scanner/analyzer"
	"diploma/scanner/saver"
	"diploma/scanner/scraper"
	"log"
)

func Scan(ctx context.Context, sc *scraper.Scraper, an *analyzer.Analyzer, sv *saver.Saver, conf scraper.Conf, workerSem chan struct{}) {
	workerSem <- struct{}{}
	go func(conf scraper.Conf) {
		defer func() {
			<-workerSem
		}()
		result := sc.Scrape(ctx, conf.Target)
		warnings := an.Analyze(ctx, conf, result)
		if len(warnings) > 0 {
			sv.Save(warnings)
		} else {
			log.Printf("Certificate for %s is OK: expires in %s, CN = %s", conf.Target, result.ExpiresIn, result.CN)
		}
	}(conf)
}
