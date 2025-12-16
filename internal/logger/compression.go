package logger

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// StartCompressionJob starts a background job to compress old log files
func StartCompressionJob(directory string, compressAfterDays int) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run once per day
		defer ticker.Stop()

		// Run immediately on start
		compressOldLogs(directory, compressAfterDays)

		// Then run daily
		for range ticker.C {
			compressOldLogs(directory, compressAfterDays)
		}
	}()
}

// compressOldLogs compresses log files older than compressAfterDays
func compressOldLogs(directory string, compressAfterDays int) {
	logger := GetLogger()

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -compressAfterDays)

	// Read directory
	entries, err := os.ReadDir(directory)
	if err != nil {
		logger.Error("Failed to read log directory for compression",
			zap.String("directory", directory),
			zap.Error(err),
		)
		return
	}

	compressedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Skip already compressed files
		if strings.HasSuffix(filename, ".gz") {
			continue
		}

		// Check if file matches log pattern
		var dateStr string
		if len(filename) >= 18 && filename[:4] == "app-" {
			dateStr = filename[4:14] // Extract "YYYY-MM-DD" from "app-YYYY-MM-DD.log"
		} else if len(filename) >= 20 && filename[:6] == "error-" {
			dateStr = filename[6:16] // Extract "YYYY-MM-DD" from "error-YYYY-MM-DD.log"
		} else {
			continue
		}

		// Parse date from filename
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// Check if file is old enough to compress
		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(directory, filename)
			compressedPath := filePath + ".gz"

			// Check if already compressed
			if _, err := os.Stat(compressedPath); err == nil {
				continue
			}

			if err := compressFile(filePath, compressedPath); err != nil {
				logger.Warn("Failed to compress log file",
					zap.String("file", filePath),
					zap.Error(err),
				)
			} else {
				compressedCount++
				logger.Info("Compressed log file",
					zap.String("file", filePath),
					zap.String("compressed", compressedPath),
					zap.Time("file_date", fileDate),
				)
			}
		}
	}

	if compressedCount > 0 {
		logger.Info("Log compression completed",
			zap.Int("compressed_files", compressedCount),
			zap.Int("compress_after_days", compressAfterDays),
		)
	}
}

// compressFile compresses a file using gzip
func compressFile(sourcePath, destPath string) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	// Set gzip header
	gzipWriter.Name = filepath.Base(sourcePath)
	gzipWriter.ModTime = sourceInfo.ModTime()

	// Copy data
	if _, err := sourceFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek source file: %w", err)
	}

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := sourceFile.Read(buf)
		if n > 0 {
			if _, writeErr := gzipWriter.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write compressed data: %w", writeErr)
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read source file: %w", err)
		}
	}

	// Flush and close gzip writer
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Close destination file
	if err := destFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	// Remove original file after successful compression
	if err := os.Remove(sourcePath); err != nil {
		// If removal fails, try to remove compressed file to avoid duplicates
		os.Remove(destPath)
		return fmt.Errorf("failed to remove original file: %w", err)
	}

	return nil
}

// CompressOldLogsNow runs compression immediately (useful for testing or manual compression)
func CompressOldLogsNow(directory string, compressAfterDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -compressAfterDays)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(filename, ".gz") {
			continue
		}

		var dateStr string
		if len(filename) >= 18 && filename[:4] == "app-" {
			dateStr = filename[4:14]
		} else if len(filename) >= 20 && filename[:6] == "error-" {
			dateStr = filename[6:16]
		} else {
			continue
		}

		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(directory, filename)
			compressedPath := filePath + ".gz"

			if _, err := os.Stat(compressedPath); err == nil {
				continue
			}

			if err := compressFile(filePath, compressedPath); err != nil {
				return fmt.Errorf("failed to compress file %s: %w", filePath, err)
			}
		}
	}

	return nil
}
