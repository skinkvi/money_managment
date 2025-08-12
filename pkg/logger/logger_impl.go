package logger

import (
	"context"
	"os"

	"github.com/skinkvi/money_managment/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	sugar *zap.SugaredLogger
}

func newZapLogger(cfg *config.LoggerConfig) (Logger, error) {
	var zapLevel zapcore.Level
	switch cfg.Level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encCfg)
	default:
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	out := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if cfg.OutputPath != "" {
		fileWS, _, err := zap.Open(cfg.OutputPath)
		if err != nil {
			return nil, err
		}
		out = []zapcore.WriteSyncer{fileWS}
	}

	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(out...), zapLevel)

	z := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return &zapLogger{sugar: z.Sugar()}, nil
}

func toZapFields(fields []Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Value)
	}

	return args
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.sugar.Debugw(msg, toZapFields(fields)...)
}
func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.sugar.Infow(msg, toZapFields(fields)...)
}
func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.sugar.Warnw(msg, toZapFields(fields)...)
}
func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.sugar.Errorw(msg, toZapFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	newSugar := l.sugar.With(toZapFields(fields)...)
	return &zapLogger{sugar: newSugar}
}
func (l *zapLogger) Sync() error {
	return l.sugar.Sync()
}
