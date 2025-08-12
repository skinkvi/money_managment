package logger

import (
	"context"

	"github.com/skinkvi/money_managment/internal/config"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)

	// возвращает новый Logger с добавленными полями.
	With(fields ...Field) Logger
	// принудительно записывает буфер (нужно вызывать перед shutdown).
	Sync() error
}

// совместима с логгером zap
type Field struct {
	Key   string
	Value interface{}
}

func New(cfg *config.LoggerConfig) (Logger, error) {
	return newZapLogger(cfg)
}
