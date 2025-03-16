package main

import (
	"context"
	"diploma/scanner/analyzer"
	"diploma/scanner/saver"
	"diploma/scanner/scanner"
	"diploma/scanner/scraper"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

	workerSem := make(chan struct{}, conf.ScraperInternalConf.PoolSize)

	var wg sync.WaitGroup

	sc := scraper.NewScraper()
	an := analyzer.NewAnalyzer()
	sv := saver.NewSaver()

	for _, cfg := range conf.Scraper {
		wg.Add(1)
		go func(conf scraper.Conf) {
			defer wg.Done()

			ticker := time.NewTicker(conf.Interval)
			defer ticker.Stop()

			scanner.Scan(ctx, sc, an, sv, conf, workerSem)

			for {
				select {
				case <-ctx.Done():
					log.Printf("Context cancelled, stopping scanning for %s", conf.Target)
					return
				case <-ticker.C:
					scanner.Scan(ctx, sc, an, sv, conf, workerSem)
				}
			}
		}(cfg)
	}

	wg.Wait()

}
