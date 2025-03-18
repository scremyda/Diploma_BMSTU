package main

import (
	"context"
	"diploma/scanner/analyzer"
	"diploma/scanner/saver"
	"diploma/scanner/scheduler"
	"diploma/scanner/scraper"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	conf, err := ReadConf(ReadArgs())
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}

	conf.AdjustScrapeIntervals()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("Received signal: %s. Initiating shutdown...", s)
		cancel()
	}()

	saver := saver.NewSaver()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, conf.Internal.PoolSize)
	for _, cfg := range conf.External {
		wg.Add(1)

		scraper := scraper.NewScraper(cfg.ScraperConf)
		analyzer := analyzer.NewAnalyzer(cfg.AnalyzerConf)

		scheduler := scheduler.NewScheduler(
			cfg.SchedulerConf,
			scraper,
			analyzer,
			saver,
		)

		go scheduler.Schedule(
			ctx,
			semaphore,
			&wg,
		)
	}

	wg.Wait()

}
