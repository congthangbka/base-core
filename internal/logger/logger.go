package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var fileWriter *DailyFileWriter
var errorFileWriter *DailyFileWriter

// Init initializes the logger with file rotation
func Init(env, logDirectory, logLevel string) error {
	// Parse log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	if env != "production" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Create main log file writer
	var err error
	fileWriter, err = NewDailyFileWriter(logDirectory, "app")
	if err != nil {
		return fmt.Errorf("failed to create file writer: %w", err)
	}

	// Create error log file writer (only errors and above)
	errorFileWriter, err = NewDailyFileWriter(logDirectory, "error")
	if err != nil {
		return fmt.Errorf("failed to create error file writer: %w", err)
	}

	// Create main file core (all logs)
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)

	// Create error file core (only errors and above)
	errorFileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(errorFileWriter),
		zapcore.ErrorLevel, // Only errors
	)

	// Create console core (for development)
	var cores []zapcore.Core
	cores = append(cores, fileCore, errorFileCore)

	if env != "production" {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Build logger with additional options
	Logger = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(0),
	)

	return nil
}

// GetLogger returns the logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Fallback to development logger if not initialized
		logger, _ := zap.NewDevelopment()
		return logger
	}
	return Logger
}

// Sync flushes the log files
func Sync() error {
	var errs []error
	if fileWriter != nil {
		if err := fileWriter.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	if errorFileWriter != nil {
		if err := errorFileWriter.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to sync log files: %v", errs)
	}
	return nil
}

// Close closes the log files
func Close() error {
	var errs []error
	if fileWriter != nil {
		if err := fileWriter.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if errorFileWriter != nil {
		if err := errorFileWriter.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close log files: %v", errs)
	}
	return nil
}

