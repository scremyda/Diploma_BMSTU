package manager

import (
	"context"
	"diploma/alerter/consumer"
	"diploma/certer/certer"
	"diploma/certer/producer"
	"diploma/certer/setter"
	"diploma/models"
	"fmt"
	"log"
	"net/url"
	"sync"
)

type Manager struct {
	consumer consumer.Interface
	producer producer.Interface
	certer   certer.Interface
	setter   setter.Interface
}

func New(
	consumer consumer.Interface,
	producer producer.Interface,
	certer certer.Interface,
	setter setter.Interface,
) *Manager {
	return &Manager{
		consumer: consumer,
		producer: producer,
		certer:   certer,
		setter:   setter,
	}
}

func (m *Manager) Manage(ctx context.Context) error {
	events, err := m.consumer.GetEvents(ctx)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for _, event := range events {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := url.Parse(event.Target)
			if err != nil {
				log.Println(fmt.Errorf("invalid URL: %w", err))
				return
			}

			cert, key, err := m.certer.GenerateCertSignedByCA(u.Scheme)
			if err != nil {
				log.Println(fmt.Errorf("failed to generate certs: %w", err))
				return
			}

			err = m.setter.Set(ctx, cert, key)
			if err != nil {
				log.Println(fmt.Errorf("failed to set certs: %w", err))
				return
			}

			event := models.ErrorEvent{}

			err = m.producer.Produce(ctx, event)
			if err != nil {
				log.Println(fmt.Errorf("failed to produce message: %w", err))
				return
			}
		}()
	}

	wg.Wait()
	return nil
}
