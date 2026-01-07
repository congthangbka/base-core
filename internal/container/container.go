package container

import (
	"llm-aggregator/internal/interfaces"
)

// ModuleContainer holds all module services for inter-module communication.
// This allows modules to call each other's functions without circular dependencies.
//
// Services are stored as specific interfaces from the interfaces package,
// providing type-safe inter-module communication.
//
// Usage in a service:
//
//	func (s *someService) SomeMethod(ctx context.Context, userID string) error {
//	    // Verify user exists
//	    if err := s.container.UserVerifier.VerifyUserExists(ctx, userID); err != nil {
//	        return err
//	    }
//	    // Get user data
//	    user, err := s.container.UserGetter.GetUserByID(ctx, userID)
//	    if err != nil {
//	        return err
//	    }
//	    // Use user data...
//	}
type ModuleContainer struct {
	// UserVerifier provides type-safe user verification across modules.
	// Use this when you only need to check if a user exists.
	UserVerifier interfaces.UserVerifier

	// UserGetter provides type-safe user data access across modules.
	// Use this when you need to get user information (name, email, status, etc.).
	UserGetter interfaces.UserGetter

	// UserService combines both UserVerifier and UserGetter for convenience.
	// Use this when you need both verification and user data.
	// This is set automatically when UserVerifier and UserGetter are the same instance.
	UserService interfaces.UserService

	// OrderService provides type-safe order access across modules.
	// Use this when you need to access order information from other modules.
	OrderService interfaces.OrderService
}

// NewModuleContainer creates a new empty module container
// Services should be set after initialization using the Set methods
// Services are initialized in router.go where we have access to repositories
func NewModuleContainer() *ModuleContainer {
	return &ModuleContainer{}
}

// SetUserVerifier sets the user verifier in the container.
// If the verifier also implements UserGetter, it will be set as UserService too.
func (c *ModuleContainer) SetUserVerifier(verifier interfaces.UserVerifier) {
	c.UserVerifier = verifier
	// If the verifier also implements UserGetter, set it as UserService too
	if getter, ok := verifier.(interfaces.UserGetter); ok {
		c.UserService = &combinedUserService{
			UserVerifier: verifier,
			UserGetter:   getter,
		}
	}
}

// SetUserGetter sets the user getter in the container.
// If the getter also implements UserVerifier, it will be set as UserService too.
func (c *ModuleContainer) SetUserGetter(getter interfaces.UserGetter) {
	c.UserGetter = getter
	// If the getter also implements UserVerifier, set it as UserService too
	if verifier, ok := getter.(interfaces.UserVerifier); ok {
		c.UserService = &combinedUserService{
			UserVerifier: verifier,
			UserGetter:   getter,
		}
	}
}

// combinedUserService combines UserVerifier and UserGetter into UserService interface
type combinedUserService struct {
	interfaces.UserVerifier
	interfaces.UserGetter
}

// SetOrderService sets the order service in the container.
// This provides type-safe inter-module access to order operations.
func (c *ModuleContainer) SetOrderService(orderService interfaces.OrderService) {
	c.OrderService = orderService
}
