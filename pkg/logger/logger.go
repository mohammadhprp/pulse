// Package logger provides logging functionality for the pulse application
package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is a singleton instance of the Zap logger
	Logger *zap.Logger
	once   sync.Once
)

// InitLogger initializes the global logger with the specified log level
// Valid levels: debug, info, warn, error, dpanic, panic, fatal
func InitLogger(level string) {
	once.Do(func() {
		// Determine the log level
		var zapLevel zapcore.Level
		switch level {
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

		// Configure logger
		config := zap.Config{
			Level:       zap.NewAtomicLevelAt(zapLevel),
			Development: false,
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			Encoding: "json",
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		var err error
		Logger, err = config.Build(zap.AddCallerSkip(1))
		if err != nil {
			// Fall back to standard logger if Zap initialization fails
			Logger = zap.NewExample()
			Logger.Error("failed to initialize zap logger", zap.Error(err))
		}

		// Sync logger when process exits
		zap.ReplaceGlobals(Logger)
	})
}

func Debug(msg string, fields ...zapcore.Field) {
	ensureLogger()
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	ensureLogger()
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	ensureLogger()
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	ensureLogger()
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	ensureLogger()
	Logger.Fatal(msg, fields...)
}

// ensureLogger makes sure the logger is initialized before use
func ensureLogger() {
	if Logger == nil {
		InitLogger("info")
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
