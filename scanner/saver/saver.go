package saver

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
)

const (
	saveErrors = ``
)

type Conf struct {
}

type Saver struct {
	db pgxtype.Querier
}

func NewSaver(db pgxtype.Querier) *Saver {
	return &Saver{
		db: db,
	}
}

func (s *Saver) Save(ctx context.Context, err error) {
	fmt.Println("[WARNING]", err)

	_, err := s.db.Exec(ctx, saveErrors)
	if err != nil {
		err = fmt.Errorf("error happened in rows.Scan: %w", err)

		return uuid.UUID{}, err
	}

}
