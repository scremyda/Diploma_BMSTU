package main

import (
	"context"
	"diploma/alerter/consumer"
	"diploma/alerter/db"
	"diploma/alerter/repo"
	"diploma/alerter/telegram"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	queueRepo := repo.NewQueue(db)

	consumer, err := consumer.New(ctx, queueRepo, conf.Consumer)
	if err != nil {
		log.Fatalf("Error create consumer: %v", err)
	}

	telegramBot, err := telegram.NewTelegramBot(consumer, conf.Telegram)
	if err != nil {
		log.Fatalf("Error create telegram bot: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		telegramBot.ProcessMessages(ctx)
	}()

	wg.Wait()
}
