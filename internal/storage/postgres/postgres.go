package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ikondratev/event-service/internal/settings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Db struct {
	Adapter *sql.DB
}

func New(ctx context.Context, settings *settings.Settings) (*Db, error) {
	db, err := sql.Open("pgx", settings.Db.Url)
	if err != nil {
		return nil, fmt.Errorf("open db error: %w", err)
	}

	db.SetConnMaxIdleTime(time.Duration(settings.Db.MaxLifeTime))
	db.SetMaxIdleConns(int(settings.Db.MaxIdleConnection))
	db.SetConnMaxLifetime(time.Duration(settings.Db.MaxLifeTime) * time.Minute)
	db.SetMaxOpenConns(settings.Db.OpenConnection)

	pingCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.Db.StartDelay)*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Db{Adapter: db}, nil
}
