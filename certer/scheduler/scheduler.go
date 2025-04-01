package scheduler

import (
	"context"
	"diploma/certer/certer"
	"diploma/certer/consumer"
	"diploma/certer/producer"
	"diploma/certer/setter"
	"diploma/models"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
)

type Config struct {
	Interval time.Duration `yaml:"interval"`
}

type Scheduler struct {
	consumer consumer.Interface
	producer producer.Interface
	certer   certer.Interface
	setter   setter.Interface
	conf     Config
}

func New(
	conf Config,
	consumer consumer.Interface,
	producer producer.Interface,
	certer certer.Interface,
	setter setter.Interface,
) *Scheduler {
	return &Scheduler{
		consumer: consumer,
		producer: producer,
		certer:   certer,
		setter:   setter,
		conf:     conf,
	}
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	ticker := time.NewTicker(s.conf.Interval)
	defer ticker.Stop()

	err := s.manage(ctx)
	if err != nil {
		log.Println(fmt.Errorf("failed to manage: %w", err))
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping scanning")
			return ctx.Err()

		case <-ticker.C:
			err := s.manage(ctx)
			if err != nil {
				log.Println(fmt.Errorf("failed to manage: %w", err))
			}
		}
	}

}

func (s *Scheduler) manage(ctx context.Context) error {
	var wg sync.WaitGroup

	events, err := s.consumer.GetEvents(ctx)
	if err != nil {
		return err
	}
	for _, event := range events {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := url.Parse(event.Target)
			if err != nil {
				log.Println(fmt.Errorf("invalid URL: %w", err))
				return
			}

			cert, key, err := s.certer.GenerateCertSignedByCA(u.Scheme)
			if err != nil {
				log.Println(fmt.Errorf("failed to generate certs: %w", err))
				return
			}

			err = s.setter.Set(ctx, u.Scheme, cert, key)
			if err != nil {
				log.Println(fmt.Errorf("failed to set certs: %w", err))
				return
			}

			event := models.AlerterEvent{
				Target:  event.Target,
				Message: fmt.Sprintf("successfully set certs for %s", u.Scheme),
			}

			err = s.producer.Produce(ctx, event)
			if err != nil {
				log.Println(fmt.Errorf("failed to produce message: %w", err))
				return
			}
		}()
	}

	wg.Wait()
	return nil
}
