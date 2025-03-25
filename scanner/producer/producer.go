package producer

import (
	"context"
	"diploma/models"
	"diploma/scanner/repo"
	"encoding/json"
	"fmt"
	"log"
)

type Interface interface {
	Produce(ctx context.Context, event models.ErrorEvent) error
}

type Config struct {
	QueueName string `yaml:"queue_name"`
	EventName string `yaml:"event_name"`
}

type Producer struct {
	conf  Config
	queue repo.Queue
}

func New(ctx context.Context, queue repo.Queue, conf Config) (*Producer, error) {
	err := queue.CreateQueue(ctx, conf.QueueName)
	if err != nil {
		return nil, fmt.Errorf("error register consumer: %w", err)
	}

	return &Producer{
		conf:  conf,
		queue: queue,
	}, nil
}

func (p *Producer) Produce(ctx context.Context, event models.ErrorEvent) error {
	eventByte, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		return err
	}

	return p.queue.Send(ctx, string(eventByte), p.conf.QueueName, p.conf.EventName)

}
