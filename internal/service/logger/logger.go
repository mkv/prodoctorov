package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger initializes a new logger instance parametrized by:
// - ctx - context.WithCancel, used to flushing any buffered log entries of parent logger
// - name - base name of logger instance
// - levelName - a minimal logging level allowed when logging a message
func NewLogger(ctx context.Context, name string, levelName string) (*zap.SugaredLogger, error) {
	level := zap.DebugLevel
	if err := level.UnmarshalText([]byte(levelName)); err != nil {
		return nil, fmt.Errorf("failed to set logging level: %w", err)
	}

	options := []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if level == zap.DebugLevel {
		options = append(options, zap.AddCaller())
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, level)
	logger := zap.New(core).WithOptions(options...)

	go func() {
		<-ctx.Done()

		_ = logger.Sync() //nolint:errcheck // silly interface of this function
	}()

	return logger.Named(name).Sugar(), nil
}
