package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CreateLogger(encoding string, level string) *zap.Logger {
	cfg := zap.Config{
		Level:            getLevel(level),
		Encoding:         encoding,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			NameKey:      "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger := zap.Must(cfg.Build())
	return logger.Named("notifications-server")
}

func getLevel(levelString string) zap.AtomicLevel {
	level := zap.NewAtomicLevel()
	switch levelString {
	case "debug":
		level.SetLevel(zapcore.DebugLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	default:
		panic(fmt.Sprintf("unknown log level %s", levelString))
	}

	return level
}
