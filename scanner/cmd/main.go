package main

import (
	"context"
	"diploma/scanner/analyzer"
	"diploma/scanner/saver"
	"diploma/scanner/scheduler"
	"diploma/scanner/scraper"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
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

	db, err := pgxpool.Connect(
		ctx,
		fmt.Sprintf(databaseConnectionStr,
			conf.Database.DBUser,
			conf.Database.DBPass,
			conf.Database.DBHost,
			conf.Database.DBPort,
			conf.Database.DBName,
		),
	)
	if err != nil {
		log.Println("failed to open postgres", err)
		return
	}
	defer db.Close()

	if err = db.Ping(ctx); err != nil {
		log.Println("failed to ping postgres", err)
		return
	}

	saver := saver.NewSaver(db)

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
