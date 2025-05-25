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

const (
	stmtName             = "insert_cert_status"
	setCertificateStatus = `
        INSERT INTO certificates_status (msg_id, payload)
        VALUES ($1, $2)
        ON CONFLICT (msg_id) DO NOTHING
    `
)

var (
	ErrFinishBatch = errors.New("error finish batch")
)

type Interface interface {
	GetAlerterEvents(ctx context.Context) ([]models.AlerterEvent, error)
	GetCerterEvents(ctx context.Context) ([]models.CerterEvent, error)
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

func (c *Consumer) GetAlerterEvents(ctx context.Context) ([]models.AlerterEvent, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, err
	}

	msgs, err := pgmq.ReadBatch(ctx, tx, c.conf.QueueName, 0, c.conf.BatchSize)
	if err != nil {
		log.Printf("Error reading message batch: %v", err)
		tx.Rollback(ctx)
		return nil, err
	}
	log.Printf("Message batch successfully read, count: %d", len(msgs))

	_, err = tx.Prepare(ctx, stmtName, setCertificateStatus)
	if err != nil {
		log.Printf("Error preparing statement %q: %v", stmtName, err)
		tx.Rollback(ctx)
		return nil, err
	}

	var events []models.AlerterEvent
	for _, m := range msgs {
		var event models.AlerterEvent
		if err := json.Unmarshal(m.Message, &event); err != nil {
			log.Printf("Error unmarshaling JSON for event %d: %v", m.MsgID, err)
			continue
		}

		if _, err := tx.Exec(ctx, stmtName, m.MsgID, m.Message); err != nil {
			log.Printf("Error executing statement %q for msg %d: %v", stmtName, m.MsgID, err)
			continue
		}

		events = append(events, event)
	}

	msgIDs := lo.Map(msgs, func(x *pgmq.Message, _ int) int64 { return x.MsgID })
	if _, err := pgmq.DeleteBatch(ctx, tx, c.conf.QueueName, msgIDs); err != nil {
		log.Printf("Error deleting message batch: %v", err)
		tx.Rollback(ctx)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	return events, nil
}

func (c *Consumer) GetCerterEvents(ctx context.Context) ([]models.CerterEvent, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return []models.CerterEvent{}, err
	}

	msgs, err := pgmq.ReadBatch(ctx, tx, c.conf.QueueName, 0, c.conf.BatchSize)
	if err != nil {
		log.Printf("Error reading message batch: %v", err)
		return []models.CerterEvent{}, err
	}
	log.Printf("Message batch successfully read, count: %d", len(msgs))

	msgsIDs := lo.Map(msgs, func(x *pgmq.Message, index int) int64 {
		return x.MsgID
	})

	_, err = pgmq.DeleteBatch(ctx, tx, c.conf.QueueName, msgsIDs)
	if err != nil {
		log.Printf("Error deleting message batch: %v", err)
		return []models.CerterEvent{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return []models.CerterEvent{}, err
	}

	var event models.CerterEvent
	events := lo.FilterMap(msgs, func(m *pgmq.Message, _ int) (models.CerterEvent, bool) {
		if err = json.Unmarshal(m.Message, &event); err != nil {
			//TODO: Need to save errors causes / success events mb
			log.Printf("Error Unmarshal JSON for event %d: %v", m.MsgID, err)
			return models.CerterEvent{}, false
		}

		return event, true
	})

	return events, nil
}
