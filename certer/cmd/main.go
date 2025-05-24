package main

import (
	"context"
	"diploma/certer/certer"
	"diploma/certer/consumer"
	"diploma/certer/db"
	"diploma/certer/producer"
	"diploma/certer/scheduler"
	"diploma/certer/setter"
	"fmt"
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

	scheduler := scheduler.New(
		conf.Scheduler,
		consumer,
		producer,
		certer,
		setter,
	)

	time.Sleep(30 * time.Second) //fix for demo

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = scheduler.Schedule(ctx)
		if err != nil {
			log.Println(fmt.Errorf("failed to schedule: %s", err))
		}
	}()

	wg.Wait()
}
