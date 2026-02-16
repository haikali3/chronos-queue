package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	cfg   zap.Config
	level zap.AtomicLevel
	log   *zap.Logger
)

func Init() {
	level = zap.NewAtomicLevel()

	env := os.Getenv("APP_ENV")
	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.Level = level
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	SetLogLevel(os.Getenv("LOG_LEVEL"))
	cfg.Level = level
	log, _ = cfg.Build()
}

func SetLogLevel(levelStr string) {
	switch strings.ToLower(levelStr) {
	case "debug":
		level.SetLevel(zapcore.DebugLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	case "warn":
		level.SetLevel(zapcore.WarnLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	default:
		log.Warn("Invalid LOG_LEVEL, defaulting to INFO", zap.String("LOG_LEVEL", levelStr))
		level.SetLevel(zapcore.InfoLevel)
	}
}

func Get() *zap.Logger {
	if log == nil {
		Init()
	}
	return log
}

func Sync() {
	if log != nil {
		log.Sync()
	}
}
