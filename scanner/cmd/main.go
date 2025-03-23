package main

import (
	"context"
	"diploma/scanner/analyzer"
	"diploma/scanner/db"
	"diploma/scanner/saver"
	"diploma/scanner/scheduler"
	"diploma/scanner/scraper"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	databaseConnectionStr string = "postgres://%v:%v@%v:%v/%v?sslmode=disable"
)

func main() {
	conf, err := ReadConf(ReadArgs())
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}

	conf.Validate()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("Received signal: %s. Initiating shutdown...", s)
		cancel()
	}()

	//time.Sleep(60 * time.Second)

	db, err := db.New(ctx, conf.Database)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	if err = db.Ping(ctx); err != nil {
		log.Println("failed to ping postgres", err)
		return
	}

	saver := saver.New(db)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, conf.Internal.PoolSize)
	for _, cfg := range conf.External {
		wg.Add(1)

		scraper := scraper.New(cfg.ScraperConf)
		analyzer := analyzer.NewAnalyzer(cfg.AnalyzerConf)

		scheduler := scheduler.New(
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
