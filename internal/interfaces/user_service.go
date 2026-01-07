package interfaces

import (
	"context"
)

// UserVerifier defines the interface for verifying user existence across modules.
// This interface is used for inter-module communication to avoid circular dependencies.
// It provides a minimal contract that modules can use to verify user existence.
//
// Usage example:
//   if err := userVerifier.VerifyUserExists(ctx, userID); err != nil {
//       return err // User not found or verification failed
//   }
type UserVerifier interface {
	// VerifyUserExists checks if a user exists by ID.
	// Returns nil if user exists, error if user not found or verification fails.
	VerifyUserExists(ctx context.Context, userID string) error
}

// UserGetter defines the interface for getting user information across modules.
// This interface provides access to user data for inter-module communication.
//
// Usage example:
//   user, err := userGetter.GetUserByID(ctx, userID)
//   if err != nil {
//       return err // User not found or retrieval failed
//   }
//   // Use user.Name, user.Email, etc.
type UserGetter interface {
	// GetUserByID retrieves user information by ID.
	// Returns user data if found, error if not found or retrieval fails.
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
}

// UserService combines UserVerifier and UserGetter for convenience.
// Modules that need both verification and user data can use this interface.
//
// Usage example:
//   // Verify and get user in one call
//   user, err := userService.GetUserByID(ctx, userID)
//   if err != nil {
//       return err // User not found
//   }
//   // Use user data
type UserService interface {
	UserVerifier
	UserGetter
}

// UserInfo contains minimal user information needed for inter-module communication.
// This avoids importing module-specific DTOs and prevents circular dependencies.
type UserInfo struct {
	ID     string // User unique identifier
	Name   string // User display name
	Email  string // User email address
	Status int    // User status (0 = inactive, 1 = active)
}

