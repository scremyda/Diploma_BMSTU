package producer

import (
	"context"
	"diploma/models"
	"encoding/json"
	"fmt"
	"github.com/craigpastro/pgmq-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type Interface interface {
	ProduceToAlerter(ctx context.Context, event models.AlerterEvent) error
	ProduceToCerter(ctx context.Context, event models.CerterEvent) error
}

type Config struct {
	AlerterQueueName string `yaml:"alerter_queue_name"`
	CerterQueueName  string `yaml:"certer_queue_name"`
}

type Producer struct {
	db   *pgxpool.Pool
	conf Config
}

func New(ctx context.Context, db *pgxpool.Pool, conf Config) (*Producer, error) {
	err := pgmq.CreateQueue(ctx, db, conf.AlerterQueueName)
	if err != nil {
		return nil, fmt.Errorf("error register alerter prosucer: %w", err)
	}

	err = pgmq.CreateQueue(ctx, db, conf.CerterQueueName)
	if err != nil {
		return nil, fmt.Errorf("error register certer prosucer: %w", err)
	}

	return &Producer{
		conf: conf,
		db:   db,
	}, nil
}

func (p *Producer) ProduceToAlerter(ctx context.Context, event models.AlerterEvent) error {
	eventByte, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Send error event", string(eventByte))

	_, err = pgmq.Send(ctx, p.db, p.conf.AlerterQueueName, eventByte)
	return err
}

func (p *Producer) ProduceToCerter(ctx context.Context, event models.CerterEvent) error {
	eventByte, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Send certer event", string(eventByte))

	_, err = pgmq.Send(ctx, p.db, p.conf.CerterQueueName, eventByte)
	return err
}
