package logger

import (
	"os"

	"assets-service/internal/ports"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogService implements Logger interface using zap.Logger
type LogService struct {
	logger *zap.Logger
}

// NewSimpleLogger creates a new zap logger adapter
func NewSimpleLogger(logger *zap.Logger) ports.Logger {
	// Add caller skip to bypass the wrapper functions
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return &LogService{
		logger: logger,
	}
}

// NewProductionZapLogger creates a production zap logger
func NewProductionZapLogger() (ports.Logger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.MessageKey = "message" // Change from "msg" to "message"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		// Remove InitialFields from here to control order
		InitialFields: nil,
	}

	zapLogger := zap.Must(config.Build())
	// Add the fields after logger creation to control order
	zapLogger = zapLogger.With(
		zap.Int("pid", os.Getpid()),
		zap.String("service", "assets-service"),
	)
	// Add caller skip to bypass the wrapper functions
	zapLogger = zapLogger.WithOptions(zap.AddCallerSkip(1))
	return &LogService{
		logger: zapLogger,
	}, nil
}

// NewDevelopmentZapLogger creates a development zap logger
func NewDevelopmentZapLogger() (ports.Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	// Add caller skip to bypass the wrapper functions
	zapLogger = zapLogger.WithOptions(zap.AddCallerSkip(1))
	return &LogService{
		logger: zapLogger,
	}, nil
}

// Info logs an info message using zap
func (l *LogService) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Sugar().Infow(msg, fields...)
	} else {
		l.logger.Info(msg)
	}
}

// Error logs an error message using zap
func (l *LogService) Error(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Sugar().Errorw(msg, fields...)
	} else {
		l.logger.Error(msg)
	}
}

// Debug logs a debug message using zap
func (l *LogService) Debug(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Sugar().Debugw(msg, fields...)
	} else {
		l.logger.Debug(msg)
	}
}

// Warn logs a warning message using zap
func (l *LogService) Warn(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Sugar().Warnw(msg, fields...)
	} else {
		l.logger.Warn(msg)
	}
}

// GetLogService returns the underlying zap.Logger (useful for components that need *zap.Logger directly)
func (l *LogService) GetLogService() *zap.Logger {
	return l.logger
}

// AsZapLogger tries to cast Logger to ZapLogger to access underlying zap.Logger
func AsLogger(logger ports.Logger) (*zap.Logger, bool) {
	if zapLogger, ok := logger.(*LogService); ok {
		return zapLogger.GetLogService(), true
	}
	return nil, false
}
