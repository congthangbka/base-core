package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

var migrations []Migration

// RegisterMigration registers a new migration
func RegisterMigration(migration Migration) {
	migrations = append(migrations, migration)
}

// RunMigrations runs all pending migrations
func RunMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(sqlDB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(sqlDB)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if isApplied(applied, migration.Version) {
			continue
		}

		// Run migration
		if err := migration.Up(db); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}

		// Record migration
		if err := recordMigration(sqlDB, migration.Version, migration.Description); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// RollbackMigration rolls back the last migration
func RollbackMigration(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Get last applied migration
	lastMigration, err := getLastMigration(sqlDB)
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	if lastMigration == "" {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find migration
	var migration *Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].Version == lastMigration {
			migration = &migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found", lastMigration)
	}

	// Rollback migration
	if err := migration.Down(db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
	}

	// Remove migration record
	if err := removeMigration(sqlDB, lastMigration); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}

// createMigrationsTable creates the migrations table
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			description VARCHAR(255),
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns list of applied migration versions
func getAppliedMigrations(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// isApplied checks if a migration version is already applied
func isApplied(applied []string, version string) bool {
	for _, v := range applied {
		if v == version {
			return true
		}
	}
	return false
}

// recordMigration records a migration as applied
func recordMigration(db *sql.DB, version, description string) error {
	query := "INSERT INTO schema_migrations (version, description) VALUES (?, ?)"
	_, err := db.Exec(query, version, description)
	return err
}

// getLastMigration returns the last applied migration version
func getLastMigration(db *sql.DB) (string, error) {
	var version string
	err := db.QueryRow("SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return version, err
}

// removeMigration removes a migration record
func removeMigration(db *sql.DB, version string) error {
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	return err
}

// LoadMigrationsFromFiles loads migrations from SQL files
func LoadMigrationsFromFiles(db *gorm.DB, migrationsDir string) error {
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	// Check if directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return nil // No migrations directory, skip
	}

	// Read migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		version := filepath.Base(file)
		version = version[:len(version)-7] // Remove .up.sql

		// Read SQL file
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		sql := string(sqlBytes)

		// Register migration
		RegisterMigration(Migration{
			Version:     version,
			Description: fmt.Sprintf("Migration from %s", file),
			Up: func(db *gorm.DB) error {
				return db.Exec(sql).Error
			},
			Down: func(db *gorm.DB) error {
				// Try to find corresponding down file
				downFile := filepath.Join(migrationsDir, version+".down.sql")
				if _, err := os.Stat(downFile); err == nil {
					downSQL, err := os.ReadFile(downFile)
					if err != nil {
						return err
					}
					return db.Exec(string(downSQL)).Error
				}
				return nil // No down migration, skip
			},
		})
	}

	return nil
}

