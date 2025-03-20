package repo

import (
	"context"
	"diploma/models"
	"encoding/json"
	"github.com/jackc/pgtype/pgxtype"
	"log"
	"time"
)

type QueueConfig struct {
	QueueName    string        `yaml:"queue_name"`
	ConsumerName string        `yaml:"consumer_name"`
	PollInterval time.Duration `yaml:"poll_interval"`
}

type QueueRepo interface {
	NextBatch(ctx context.Context, queueName, consumerName string) (int64, error)
	GetBatchEvents(ctx context.Context, batchID int64) ([]models.ErrorEvent, error)
	FinishBatch(ctx context.Context, queueName, consumerName string, batchID int64) error
}

type PGQRepo struct {
	db pgxtype.Querier
}

func New(db pgxtype.Querier) *PGQRepo {
	return &PGQRepo{
		db: db,
	}
}

func (r *PGQRepo) NextBatch(ctx context.Context, queueName, consumerName string) (int64, error) {
	var batchID int64
	err := r.db.QueryRow(ctx, "SELECT pgq.next_batch($1, $2)", queueName, consumerName).Scan(&batchID)
	if err != nil {
		return 0, err
	}
	return batchID, nil
}

func (r *PGQRepo) GetBatchEvents(ctx context.Context, batchID int64) ([]models.ErrorEvent, error) {
	rows, err := r.db.Query(ctx, "SELECT ev_id, ev_type, ev_data FROM pgq.get_batch_events($1)", batchID)
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
			log.Printf("Ошибка сканирования события: %v", err)
			continue
		}

		var event models.ErrorEvent
		if err := json.Unmarshal([]byte(evData), &event); err != nil {
			log.Printf("Ошибка декодирования JSON для события %d: %v", evID, err)
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *PGQRepo) FinishBatch(ctx context.Context, queueName, consumerName string, batchID int64) error {
	_, err := r.db.Exec(ctx, "SELECT pgq.finish_batch($1, $2, $3)", queueName, consumerName, batchID)
	return err
}
