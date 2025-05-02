package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// логгер - zap
type Logger = *zap.Logger

// кастомный формат времени
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	formated := t.Format("2006-01-02 15:04.05.000 MST")
	enc.AppendString(formated)
}

func NewLogger() Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = CustomTimeEncoder
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	logger, err := cfg.Build()
	if err != nil {
		logger.Error("failed to build zap logger, usuing fallback logger:", zap.Error(err))
		// запуск фейк-логгера
		return zap.NewNop()
	}
	//nolint:errcheck
	defer logger.Sync()

	return logger
}
