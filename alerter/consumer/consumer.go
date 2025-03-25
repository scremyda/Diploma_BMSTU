package consumer

import (
	"context"
	"diploma/alerter/repo"
	"diploma/models"
	"errors"
	"fmt"
	"log"
)

var (
	ErrFinishBatch = errors.New("error finish batch")
)

type Interface interface {
	GetEvents(ctx context.Context) ([]models.ErrorEvent, error)
}

type Config struct {
	QueueName    string `yaml:"queue_name"`
	ConsumerName string `yaml:"consumer_name"`
}

type Consumer struct {
	queue repo.Queue
	conf  Config
}

func New(ctx context.Context, queue repo.Queue, conf Config) (*Consumer, error) {
	err := queue.RegisterConsumer(ctx, conf.QueueName, conf.ConsumerName)
	if err != nil {
		return nil, fmt.Errorf("error register consumer: %w", err)
	}

	return &Consumer{
		queue: queue,
		conf:  conf,
	}, nil
}

func (c *Consumer) GetEvents(ctx context.Context) ([]models.ErrorEvent, error) {
	batchID, err := c.queue.NextBatch(ctx, c.conf.QueueName, c.conf.ConsumerName)
	if err != nil {
		log.Printf("Error receiving new batch: %v", err)
		return nil, err
	}
	if batchID <= 0 {
		return nil, errors.New("no batch found")
	}

	events, err := c.queue.GetBatchEvents(ctx, batchID)
	if err != nil {
		log.Printf("Error getting batch events for batch %d: %v", batchID, err)
		return nil, err
	}

	if err := c.queue.FinishBatch(ctx, batchID); err != nil {
		log.Printf("Error finishing batch %d: %v", batchID, err)
		return events, ErrFinishBatch
	}

	log.Printf("Batch %d processed", batchID)

	return events, err
}
