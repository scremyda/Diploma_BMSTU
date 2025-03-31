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
	Produce(ctx context.Context, event models.AlerterEvent) error
}

type Config struct {
	QueueName string `yaml:"queue_name"`
}

type Producer struct {
	db   *pgxpool.Pool
	conf Config
}

func New(ctx context.Context, db *pgxpool.Pool, conf Config) (*Producer, error) {
	err := pgmq.CreateQueue(ctx, db, conf.QueueName)
	if err != nil {
		return nil, fmt.Errorf("error register consumer: %w", err)
	}

	return &Producer{
		conf: conf,
		db:   db,
	}, nil
}

func (p *Producer) Produce(ctx context.Context, event models.AlerterEvent) error {
	eventByte, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Send error event", string(eventByte))

	_, err = pgmq.Send(ctx, p.db, p.conf.QueueName, eventByte)
	return err
}
