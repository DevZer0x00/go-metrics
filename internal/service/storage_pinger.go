package service

import (
	"context"
	"database/sql"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type NopPinger struct{}

func (p *NopPinger) Ping(ctx context.Context) error {
	return nil
}

type DBPinger struct {
	db *sql.DB
}

func (p *DBPinger) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func NewDBPinger(db *sql.DB) *DBPinger {
	return &DBPinger{db: db}
}
