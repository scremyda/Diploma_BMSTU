package scheduler

import (
	"context"
	"diploma/models"
	"diploma/scanner/analyzer"
	"diploma/scanner/repo"
	"diploma/scanner/scraper"
	"encoding/json"
	"log"
	"math/rand/v2"
	"sync"
	"time"
)

type Conf struct {
	Interval    time.Duration `yaml:"timeout"`
	RangeFactor float64       `yaml:"range_factor"`
}

type Scheduler struct {
	conf     Conf
	scraper  *scraper.Scraper
	analyzer *analyzer.Analyzer
	saver    *repo.Repo
}

func New(
	cfg Conf,
	scraper *scraper.Scraper,
	analyzer *analyzer.Analyzer,
	saver *repo.Repo,
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

	ticker := time.NewTicker(s.randomizeInterval())
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

func Scan(ctx context.Context, sc *scraper.Scraper, an *analyzer.Analyzer, sv *repo.Repo, semaphore chan struct{}) {
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
			event := models.ErrorEvent{
				Target:  scrapeInfo.Target,
				Message: err.Error(),
			}

			eventByte, err := json.Marshal(event)
			if err != nil {
				log.Println(err)
				return
			}

			if err := sv.Send(ctx, string(eventByte)); err != nil {
				log.Println("save error: ", err)
			}
		} else {
			log.Printf("Certificate for %s is OK: expires in %s, CN = %s", scrapeInfo.Target, scrapeInfo.ExpiresIn, scrapeInfo.CN)
		}
	}()
}

func (s *Scheduler) randomizeInterval() time.Duration {
	if s.conf.Interval <= 0 {
		return s.conf.Interval
	}

	band := float64(s.conf.Interval) * s.conf.RangeFactor
	offset := rand.Float64()*2*band - band // [-band, +band)

	return s.conf.Interval + time.Duration(offset)
}
