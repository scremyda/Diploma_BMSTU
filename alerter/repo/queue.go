package repo

import (
	"context"
	"diploma/models"
	"encoding/json"
	"log"

	"github.com/jackc/pgtype/pgxtype"
)

const (
	nextBatch        = "SELECT * FROM pgq.next_batch($1, $2)"
	getBatchEvents   = "SELECT ev_id, ev_type, ev_data FROM pgq.get_batch_events($1)"
	finishBatch      = "SELECT * FROM pgq.finish_batch($1)"
	registerConsumer = "select * from pgq.register_consumer($1, $2);"
)

type Queue interface {
	NextBatch(ctx context.Context, queueName, consumerName string) (int64, error)
	GetBatchEvents(ctx context.Context, batchID int64) ([]models.ErrorEvent, error)
	FinishBatch(ctx context.Context, batchID int64) error
	RegisterConsumer(ctx context.Context, queueName, consumerName string) error
}

type Repo struct {
	db pgxtype.Querier
}

func NewQueue(db pgxtype.Querier) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) NextBatch(ctx context.Context, queueName, consumerName string) (int64, error) {
	var batchID int64
	err := r.db.QueryRow(
		ctx,
		nextBatch,
		queueName,
		consumerName,
	).Scan(&batchID)
	if err != nil {
		return 0, err
	}

	return batchID, nil
}

func (r *Repo) GetBatchEvents(ctx context.Context, batchID int64) ([]models.ErrorEvent, error) {
	rows, err := r.db.Query(
		ctx,
		getBatchEvents,
		batchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.ErrorEvent
	for rows.Next() {
		var evID int64
		var evType string
		var evData string
		if err := rows.Scan(&evID, &evType, &evData); err != nil {
			log.Printf("Error event scan: %v", err)
			continue
		}

		var event models.ErrorEvent
		if err := json.Unmarshal([]byte(evData), &event); err != nil {
			log.Printf("Error Unmarshal JSON for event %d: %v", evID, err)
			continue
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *Repo) FinishBatch(ctx context.Context, batchID int64) error {
	_, err := r.db.Exec(
		ctx,
		finishBatch,
		batchID,
	)

	return err
}

func (r *Repo) RegisterConsumer(ctx context.Context, queueName, consumerName string) error {
	_, err := r.db.Exec(
		ctx,
		registerConsumer,
		queueName,
		consumerName,
	)

	return err
}
