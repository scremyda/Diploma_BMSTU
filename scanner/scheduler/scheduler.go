package scheduler

import (
	"context"
	"diploma/models"
	"diploma/scanner/analyzer"
	"diploma/scanner/producer"
	"diploma/scanner/scraper"
	"log"
	"math/rand/v2"
	"sync"
	"time"
)

type Conf struct {
	Interval    time.Duration `yaml:"interval"`
	RangeFactor float64       `yaml:"range_factor"`
}

type Scheduler struct {
	conf     Conf
	scraper  *scraper.Scraper
	analyzer *analyzer.Analyzer
	producer producer.Interface
}

func New(
	cfg Conf,
	scraper *scraper.Scraper,
	analyzer *analyzer.Analyzer,
	producer producer.Interface,
) *Scheduler {
	return &Scheduler{
		conf:     cfg,
		scraper:  scraper,
		analyzer: analyzer,
		producer: producer,
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

	s.scan(ctx, semaphore)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping scanning")
			return ctx.Err()

		case <-ticker.C:
			s.scan(ctx, semaphore)
		}
	}
}

func (s *Scheduler) scan(
	ctx context.Context,
	semaphore chan struct{},
) {
	semaphore <- struct{}{}
	go func() {
		defer func() {
			<-semaphore
		}()
		scrapeInfo, err := s.scraper.Scrape(ctx)
		if err != nil {
			log.Println("Scraper error: ", err)
			return
		}

		errAnalyzer := s.analyzer.Analyze(ctx, scrapeInfo)
		if errAnalyzer != nil {
			certerEvent := models.CerterEvent{
				Target: scrapeInfo.Target,
			}
			err := s.producer.ProduceToCerter(ctx, certerEvent)
			if err != nil {
				log.Println("Producer error: ", err)
				return
			}

			alerterEvent := models.AlerterEvent{
				Target:  scrapeInfo.Target,
				Message: errAnalyzer.Error(),
			}
			err = s.producer.ProduceToAlerter(ctx, alerterEvent)
			if err != nil {
				log.Println("Producer error: ", err)
				return
			}
		} else {
			log.Printf(
				"Certificate for %s is OK: expires in %s, CN = %s",
				scrapeInfo.Target,
				scrapeInfo.ExpiresIn,
				scrapeInfo.CN,
			)
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
