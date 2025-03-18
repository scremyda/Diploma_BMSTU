package scheduler

import (
	"context"
	"diploma/scanner/analyzer"
	"diploma/scanner/saver"
	"diploma/scanner/scraper"
	"log"
	"sync"
	"time"
)

type Conf struct {
	Interval time.Duration `yaml:"timeout"`
}

type Scheduler struct {
	conf     Conf
	scraper  *scraper.Scraper
	analyzer *analyzer.Analyzer
	saver    *saver.Saver
}

func NewScheduler(
	cfg Conf,
	scraper *scraper.Scraper,
	analyzer *analyzer.Analyzer,
	saver *saver.Saver,
) *Scheduler {
	return &Scheduler{
		conf:     cfg,
		scraper:  scraper,
		analyzer: analyzer,
		saver:    saver,
	}
}

func (s *Scheduler) Schedule(
	ctx context.Context,
	semaphore chan struct{},
	wg *sync.WaitGroup,
) error {
	defer wg.Done()

	ticker := time.NewTicker(s.conf.Interval)
	defer ticker.Stop()

	Scan(ctx, s.scraper, s.analyzer, s.saver, semaphore)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping scanning")
			return ctx.Err()

		case <-ticker.C:
			Scan(ctx, s.scraper, s.analyzer, s.saver, semaphore)
		}
	}
}

func Scan(ctx context.Context, sc *scraper.Scraper, an *analyzer.Analyzer, sv *saver.Saver, semaphore chan struct{}) {
	semaphore <- struct{}{}
	go func() {
		defer func() {
			<-semaphore
		}()
		scrapeInfo, err := sc.Scrape(ctx)
		if err != nil {
			log.Println("Scraper error: ", err)
			return
		}

		err = an.Analyze(ctx, scrapeInfo)
		if err != nil {
			sv.Save(err)
		} else {
			log.Printf("Certificate for %s is OK: expires in %s, CN = %s", scrapeInfo.Target, scrapeInfo.ExpiresIn, scrapeInfo.CN)
		}
	}()
}
