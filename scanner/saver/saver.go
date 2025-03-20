package saver

import (
	"context"
	"diploma/models"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
)

const saveErrors = "SELECT pgq.insert_event('error_queue', 'error', $1)"

type Saver struct {
	db pgxtype.Querier
}

func New(db pgxtype.Querier) *Saver {
	return &Saver{
		db: db,
	}
}

func (s *Saver) Save(ctx context.Context, event models.ErrorEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	fmt.Println("[WARNING]", string(data))

	_, err = s.db.Exec(ctx, saveErrors, string(data))
	return err
}
