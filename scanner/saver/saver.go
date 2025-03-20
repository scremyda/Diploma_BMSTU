package saver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
)

const saveErrors = "SELECT pgq.insert_event('error_queue', 'error', $1)"

type Saver struct {
	db pgxtype.Querier
}

func NewSaver(db pgxtype.Querier) *Saver {
	return &Saver{
		db: db,
	}
}

type ErrorEvent struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}

func (s *Saver) Save(ctx context.Context, event ErrorEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	fmt.Println("[WARNING]", string(data))

	_, err = s.db.Exec(ctx, saveErrors, string(data))
	return err
}
