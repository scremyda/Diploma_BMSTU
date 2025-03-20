package main

import (
	"context"
	"diploma/alerter/repo"
	"diploma/alerter/telegram"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	databaseConnectionStr string = "postgres://%v:%v@%v:%v/%v?sslmode=disable"
)

func main() {
	conf, err := ReadConf(ReadArgs())
	if err != nil {
		log.Fatalf("Ошибка чтения конфига: %v", err)
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

	queueRepo := repo.New(db)

	telegramToken := conf.Telegram.BotToken
	telegramChatID := conf.Telegram.ChatID
	queueName := "error_queue"
	consumerName := "telegram_bot_consumer"

	telegramBot, err := telegram.NewTelegramBot(queueRepo, telegramToken, telegramChatID, queueName, consumerName)
	if err != nil {
		log.Fatalf("Ошибка создания Telegram бота: %v", err)
	}

	go telegramBot.ProcessQueue(ctx, 5*time.Second)

	select {}
}
