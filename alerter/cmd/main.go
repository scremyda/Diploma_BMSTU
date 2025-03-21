package main

import (
	"context"
	"diploma/alerter/db"
	"diploma/alerter/repo"
	"diploma/alerter/telegram"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	conf, err := ReadConf(ReadArgs())
	if err != nil {
		log.Fatalf("Config read error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("Received signal: %s. Initiating shutdown...", s)
		cancel()
	}()

	db, err := db.New(ctx, conf.Database)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	queueRepo := repo.New(db, conf.Queue)

	telegramBot, err := telegram.NewTelegramBot(queueRepo, conf.Telegram)
	if err != nil {
		log.Fatalf("Error create telegram bot: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		telegramBot.ProcessQueue(ctx, conf.Queue.PollInterval)
	}()

	wg.Wait()
}
