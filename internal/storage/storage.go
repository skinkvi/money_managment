package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skinkvi/money_managment/internal/config"
	"github.com/skinkvi/money_managment/pkg/logger"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrDB                = errors.New("database error")
	ErrNoUsers           = errors.New("no users found")
)

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close()
}

type DB struct {
	Pool DBPool
	log  logger.Logger
}

func Connect(ctx context.Context, cfg config.DBConfig, log logger.Logger) (*DB, error) {
	poolCfg := pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{},
	}
	// пришлось составлять конфиг так, потому что если я пытаюсь его составить внутри струтктуры ConnConfig{} то пишет что нет ни одного из перечисленных полей
	poolCfg.MaxConns = int32(cfg.MaxConn)
	poolCfg.ConnConfig.Host = cfg.Host
	poolCfg.ConnConfig.Port = uint16(cfg.Port)
	poolCfg.ConnConfig.User = cfg.User
	poolCfg.ConnConfig.Password = cfg.Password
	poolCfg.ConnConfig.Database = cfg.DBName

	pool, err := pgxpool.NewWithConfig(context.Background(), &poolCfg)
	if err != nil {
		log.Error(ctx, "cannot create pool with config")
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &DB{Pool: pool, log: log}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
