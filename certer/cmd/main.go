package main

import (
	"context"
	"diploma/certer/consumer"
	"diploma/certer/db"
	"diploma/certer/producer"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"os"
	"os/signal"
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

	consumer := consumer.New(db, conf.Consumer)
	producer, err := producer.New(ctx, db, conf.Producer)
	if err != nil {
		log.Println(err)
		return
	}
}
