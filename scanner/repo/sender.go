package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype/pgxtype"
)

const saveErrors = "SELECT * FROM pgq.insert_event($1, $2, $3)"

type Config struct {
	QueueName string `yaml:"queue_name"`
	EventName string `yaml:"event_name"`
}

type Repo struct {
	db   pgxtype.Querier
	conf Config
}

func New(db pgxtype.Querier, conf Config) *Repo {
	return &Repo{
		db:   db,
		conf: conf,
	}
}

func (s *Repo) Send(ctx context.Context, event string) error {
	fmt.Println("[WARNING]", event)

	_, err := s.db.Exec(
		ctx,
		saveErrors,
		s.conf.QueueName,
		s.conf.EventName,
		event,
	)
	return err
}
