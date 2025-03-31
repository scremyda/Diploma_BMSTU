package main

import (
	"context"
	"diploma/certer/certer"
	"diploma/certer/consumer"
	"diploma/certer/db"
	"diploma/certer/manager"
	"diploma/certer/producer"
	"diploma/certer/setter"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	certer := certer.New(conf.Certer)
	setter := setter.New(conf.Setter)

	manager := manager.New(
		consumer,
		producer,
		certer,
		setter,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err = manager.Manage(ctx)
			if err != nil {
				log.Println(fmt.Errorf("failes to manage: %s", err))
			}
			time.Sleep(5 * time.Second) //TODO: move to config / fix logic
		}
	}()

	wg.Wait()
}
