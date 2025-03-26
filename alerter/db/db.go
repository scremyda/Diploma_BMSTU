package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	databaseConnectionStr string = "postgres://%v:%v@%v:%v/%v?sslmode=disable"
)

type Config struct {
	DBUser string `yaml:"db_user"`
	DBPass string `yaml:"db_pass"`
	DBHost string `yaml:"db_host"`
	DBPort int    `yaml:"db_port"`
	DBName string `yaml:"db_name"`
}

func New(ctx context.Context, conf Config) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(
		ctx,
		fmt.Sprintf(databaseConnectionStr,
			conf.DBUser,
			conf.DBPass,
			conf.DBHost,
			conf.DBPort,
			conf.DBName,
		),
	)
	if err != nil {
		log.Println("failed to open postgres", err)
		return nil, err
	}

	if err = db.Ping(ctx); err != nil {
		log.Println("failed to ping postgres", err)
		return nil, err
	}

	return db, nil
}
