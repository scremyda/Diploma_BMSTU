package consumer

import (
	"context"
	"diploma/models"
	"encoding/json"
	"errors"
	"github.com/craigpastro/pgmq-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
	"log"
)

var (
	ErrFinishBatch = errors.New("error finish batch")
)

type Interface interface {
	GetEvents(ctx context.Context) ([]models.ErrorEvent, error)
}

type Config struct {
	QueueName string `yaml:"queue_name"`
	BatchSize int64  `yaml:"batch_size"`
}

type Consumer struct {
	db   *pgxpool.Pool
	conf Config
}

func New(db *pgxpool.Pool, conf Config) *Consumer {
	return &Consumer{
		db:   db,
		conf: conf,
	}
}

func (c *Consumer) GetEvents(ctx context.Context) ([]models.ErrorEvent, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return []models.ErrorEvent{}, err
	}

	msgs, err := pgmq.ReadBatch(ctx, tx, c.conf.QueueName, 0, c.conf.BatchSize)
	if err != nil {
		log.Printf("Error reading message batch: %v", err)
		return []models.ErrorEvent{}, err
	}
	log.Printf("Message batch successfully read, count: %d", len(msgs))

	msgsIDs := lo.Map(msgs, func(x *pgmq.Message, index int) int64 {
		return x.MsgID
	})

	_, err = pgmq.DeleteBatch(ctx, tx, c.conf.QueueName, msgsIDs)
	if err != nil {
		log.Printf("Error deleting message batch: %v", err)
		return []models.ErrorEvent{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return []models.ErrorEvent{}, err
	}

	var event models.ErrorEvent
	events := lo.FilterMap(msgs, func(m *pgmq.Message, _ int) (models.ErrorEvent, bool) {
		if err = json.Unmarshal(m.Message, &event); err != nil {
			//TODO: Need to save errors causes / success events mb
			log.Printf("Error Unmarshal JSON for event %d: %v", m.MsgID, err)
			return models.ErrorEvent{}, false
		}

		return event, true
	})

	return events, nil
}
