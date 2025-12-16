package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// StartCleanupJob starts a background job to clean up old log files
func StartCleanupJob(directory string, retentionDays int) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run once per day
		defer ticker.Stop()

		// Run immediately on start
		cleanupOldLogs(directory, retentionDays)

		// Then run daily
		for range ticker.C {
			cleanupOldLogs(directory, retentionDays)
		}
	}()
}

// cleanupOldLogs removes log files older than retentionDays
func cleanupOldLogs(directory string, retentionDays int) {
	logger := GetLogger()

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Read directory
	entries, err := os.ReadDir(directory)
	if err != nil {
		logger.Error("Failed to read log directory",
			zap.String("directory", directory),
			zap.Error(err),
		)
		return
	}

	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		
		// Check if file matches log pattern (app-YYYY-MM-DD.log, error-YYYY-MM-DD.log, or compressed .gz files)
		var dateStr string
		isCompressed := strings.HasSuffix(filename, ".gz")
		
		if isCompressed {
			// Handle compressed files: app-YYYY-MM-DD.log.gz or error-YYYY-MM-DD.log.gz
			baseName := strings.TrimSuffix(filename, ".gz")
			if len(baseName) >= 18 && baseName[:4] == "app-" {
				dateStr = baseName[4:14]
			} else if len(baseName) >= 20 && baseName[:6] == "error-" {
				dateStr = baseName[6:16]
			} else {
				continue
			}
		} else {
			// Handle uncompressed files
			if len(filename) >= 18 && filename[:4] == "app-" {
				dateStr = filename[4:14] // Extract "YYYY-MM-DD" from "app-YYYY-MM-DD.log"
			} else if len(filename) >= 20 && filename[:6] == "error-" {
				dateStr = filename[6:16] // Extract "YYYY-MM-DD" from "error-YYYY-MM-DD.log"
			} else {
				continue
			}
		}

		// Parse date from filename
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// Check if file is older than retention period
		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(directory, filename)
			if err := os.Remove(filePath); err != nil {
				logger.Warn("Failed to delete old log file",
					zap.String("file", filePath),
					zap.Error(err),
				)
			} else {
				deletedCount++
				logger.Info("Deleted old log file",
					zap.String("file", filePath),
					zap.Time("file_date", fileDate),
				)
			}
		}
	}

	if deletedCount > 0 {
		logger.Info("Log cleanup completed",
			zap.Int("deleted_files", deletedCount),
			zap.Int("retention_days", retentionDays),
		)
	}
}

// CleanupOldLogsNow runs cleanup immediately (useful for testing or manual cleanup)
func CleanupOldLogsNow(directory string, retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		var dateStr string
		isCompressed := strings.HasSuffix(filename, ".gz")
		
		if isCompressed {
			baseName := strings.TrimSuffix(filename, ".gz")
			if len(baseName) >= 18 && baseName[:4] == "app-" {
				dateStr = baseName[4:14]
			} else if len(baseName) >= 20 && baseName[:6] == "error-" {
				dateStr = baseName[6:16]
			} else {
				continue
			}
		} else {
			if len(filename) >= 18 && filename[:4] == "app-" {
				dateStr = filename[4:14]
			} else if len(filename) >= 20 && filename[:6] == "error-" {
				dateStr = filename[6:16]
			} else {
				continue
			}
		}

		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(directory, filename)
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

