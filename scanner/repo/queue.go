package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype/pgxtype"
)

const (
	saveErrors  = "SELECT * FROM pgq.insert_event($1, $2, $3)"
	createQueue = "SELECT * FROM pgq.create_queue($1);"
)

type Queue interface {
	Send(ctx context.Context, event, queueName, eventName string) error
	CreateQueue(ctx context.Context, queueName string) error
}

type Repo struct {
	db pgxtype.Querier
}

func NewQueue(db pgxtype.Querier) *Repo {
	return &Repo{
		db: db,
	}
}

func (s *Repo) Send(ctx context.Context, event, queueName, eventName string) error {
	fmt.Println("[WARNING]", event)

	_, err := s.db.Exec(
		ctx,
		saveErrors,
		queueName,
		eventName,
		event,
	)
	return err
}

func (s *Repo) CreateQueue(ctx context.Context, queueName string) error {
	_, err := s.db.Exec(
		ctx,
		createQueue,
		queueName,
	)
	return err
}
