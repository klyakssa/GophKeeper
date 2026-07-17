package logger

import (
	"gophkeeper/internal/config"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new instance of Logger
func NewLogger(cfg *config.LoggingConfiguration, name string) *Logger {
	logger := zap.New(configure(cfg, name), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return &Logger{Logger: logger}
}

func configure(cfg *config.LoggingConfiguration, name string) zapcore.Core {
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Path, name+".log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
	})

	var priority zap.LevelEnablerFunc
	switch cfg.Level {
	case "debug":
		priority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zap.DebugLevel
		})
	case "info":
		priority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zap.InfoLevel
		})
	case "warn":
		priority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zap.WarnLevel
		})
	case "error":
		priority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zap.ErrorLevel
		})
	case "fatal":
		priority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zap.FatalLevel
		})
	default:
		priority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return true
		})
	}

	enconfig := zap.NewProductionEncoderConfig()
	enconfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	enconfig.TimeKey = "timestamp"

	consoleWriter := zapcore.Lock(os.Stdout)
	jsonEncoder := zapcore.NewJSONEncoder(enconfig)
	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, priority),
		zapcore.NewCore(jsonEncoder, fileWriter, priority),
	)
}
