package common

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// TransactionManager provides transaction management utilities.
// This helps services execute multiple repository operations atomically.
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager.
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Execute executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// Usage:
//   err := txManager.Execute(ctx, func(tx *gorm.DB) error {
//       // Perform multiple operations
//       if err := repo1.WithTx(tx).Create(ctx, entity1); err != nil {
//           return err
//       }
//       if err := repo2.WithTx(tx).Update(ctx, entity2); err != nil {
//           return err
//       }
//       return nil
//   })
func (tm *TransactionManager) Execute(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(fn)
}

// Transaction is a convenience function that executes a function within a transaction.
// This is a standalone function that can be used without creating a TransactionManager.
//
// Usage:
//   err := common.Transaction(db, func(tx *gorm.DB) error {
//       // Perform operations
//       return nil
//   })
func Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// TransactionWithContext executes a function within a transaction with context.
//
// Usage:
//   err := common.TransactionWithContext(ctx, db, func(tx *gorm.DB) error {
//       // Perform operations
//       return nil
//   })
func TransactionWithContext(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}

// IsTransactionError checks if an error is a transaction-related error.
func IsTransactionError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common transaction errors
	return errors.Is(err, gorm.ErrInvalidTransaction) ||
		errors.Is(err, gorm.ErrNotImplemented) ||
		errors.Is(err, gorm.ErrMissingWhereClause)
}

