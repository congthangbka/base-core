# Repository Pattern vs Direct Store Usage - Comparison

## Overview

This document compares two approaches for database operations:
1. **Direct Store Usage**: Using `store.ModifyRecord` directly in services
2. **Repository Pattern**: Wrapping DB operations in repository methods

## Approach 1: Direct Store Usage in Service

### Example Code
```go
// In service layer
if err := store.ModifyRecord[entity.DomainGroupHTMLTemplate](
    map[string]interface{}{"id": req.Id}, 
    dataUpdate,
); err != nil {
    return common.InternalError(err)
}
```

### Advantages ✅
- **Quick to implement**: Less code, faster development for simple cases
- **No extra abstraction**: Direct access to generic store functions
- **Suitable for prototypes**: Good for rapid prototyping or simple CRUD

### Disadvantages ❌
- **Violates Clean Architecture**: Service layer directly depends on data access layer
- **Hard to test**: Must mock `store` package instead of repository interface
- **Difficult to extend**: Adding validation, transformation, or business logic requires changes in multiple places
- **Transaction management**: No easy way to handle transactions (no `WithTx` pattern)
- **No abstraction**: Changing ORM or database requires modifying service code
- **Code duplication**: Similar DB operations scattered across services
- **Maintenance burden**: Business logic mixed with data access logic
- **Type safety**: Less type-safe compared to repository methods

## Approach 2: Repository Pattern (Recommended)

### Example Code
```go
// In repository layer
type DomainGroupRepository interface {
    Update(ctx context.Context, id string, data map[string]interface{}) error
    WithTx(tx *gorm.DB) DomainGroupRepository
}

func (r *domainGroupRepository) Update(ctx context.Context, id string, data map[string]interface{}) error {
    result := r.db.WithContext(ctx).
        Model(&entity.DomainGroupHTMLTemplate{}).
        Where("id = ?", id).
        Updates(data)
    if result.Error != nil {
        return common.WrapError(result.Error, "failed to update domain group")
    }
    if result.RowsAffected == 0 {
        return common.ErrNotFound
    }
    return nil
}

// In service layer
if err := r.repo.Update(ctx, req.Id, dataUpdate); err != nil {
    if errors.Is(err, common.ErrNotFound) {
        return common.NewServiceError(err, "Domain group not found", common.ErrorCodeNotFound)
    }
    return common.InternalError(err)
}
```

### Advantages ✅
- **Clean Architecture compliance**: Proper separation of concerns
- **Easy to test**: Mock repository interface in unit tests
- **Easy to extend**: Add validation, transformation, or business logic in one place
- **Transaction support**: Can use `WithTx` pattern for complex operations
- **Abstraction**: Changing ORM or database doesn't affect service layer
- **Code reusability**: DB logic centralized in repository
- **Maintainability**: Business logic separated from data access
- **Type safety**: Strong typing with repository interfaces
- **Consistent error handling**: Standardized error wrapping and transformation

### Disadvantages ❌
- **More code**: Requires additional repository methods
- **Slightly more setup**: Need to define repository interface and implementation

## Real-World Scenarios

### Scenario 1: Simple Update Operation

**Direct Store:**
```go
// Service
if err := store.ModifyRecord[entity.User](map[string]interface{}{"id": id}, updates); err != nil {
    return err
}
```

**Repository:**
```go
// Repository
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
    result := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", user.ID).Updates(user)
    if result.Error != nil {
        return common.WrapError(result.Error, "failed to update user")
    }
    if result.RowsAffected == 0 {
        return common.ErrNotFound
    }
    return nil
}

// Service
if err := s.repo.Update(ctx, user); err != nil {
    return err
}
```

### Scenario 2: Complex Operation with Validation

**Direct Store (Problematic):**
```go
// Service - validation mixed with DB access
if err := validateUpdate(dataUpdate); err != nil {
    return err
}
if err := store.ModifyRecord[entity.User](map[string]interface{}{"id": id}, dataUpdate); err != nil {
    return err
}
// Where to add audit logging? Where to add caching? Where to add transaction?
```

**Repository (Clean):**
```go
// Repository - all DB concerns in one place
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
    // Validation, transformation, audit logging can all be here
    result := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", user.ID).Updates(user)
    // ... error handling
    return nil
}

// Service - pure business logic
if err := s.repo.Update(ctx, user); err != nil {
    return err
}
```

### Scenario 3: Testing

**Direct Store:**
```go
// Hard to test - must mock store package
func TestUserService_Update(t *testing.T) {
    // How to mock store.ModifyRecord? Need to use dependency injection or global mocks
}
```

**Repository:**
```go
// Easy to test - mock repository interface
func TestUserService_Update(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    // Test service logic with mocked repository
}
```

## Recommendation

**Use Repository Pattern (Approach 2)** for production code because:

1. **Maintainability**: Code is easier to maintain and understand
2. **Testability**: Unit tests are simpler and more reliable
3. **Scalability**: Easy to add features like caching, audit logging, etc.
4. **Consistency**: Matches your existing codebase pattern (see `user_repository.go`)
5. **Future-proof**: Easy to migrate to different database or ORM

**Consider Direct Store** only for:
- Quick prototypes or proof-of-concept
- Very simple, one-off operations that won't be reused
- Internal tools with minimal requirements

## Current Codebase Pattern

Your codebase already follows the Repository Pattern (see `internal/modules/user/repository/user_repository.go`). For consistency and maintainability, continue using this pattern for all modules.

## Implementation Template

When adding a new entity, follow this pattern:

```go
// 1. Define repository interface
type DomainGroupRepository interface {
    Update(ctx context.Context, id string, data map[string]interface{}) error
    WithTx(tx *gorm.DB) DomainGroupRepository
}

// 2. Implement repository
type domainGroupRepository struct {
    db *gorm.DB
}

func NewDomainGroupRepository(db *gorm.DB) DomainGroupRepository {
    return &domainGroupRepository{db: db}
}

func (r *domainGroupRepository) Update(ctx context.Context, id string, data map[string]interface{}) error {
    result := r.db.WithContext(ctx).
        Model(&entity.DomainGroupHTMLTemplate{}).
        Where("id = ?", id).
        Updates(data)
    if result.Error != nil {
        return common.WrapError(result.Error, "failed to update domain group")
    }
    if result.RowsAffected == 0 {
        return common.ErrNotFound
    }
    return nil
}

// 3. Use in service
type domainGroupService struct {
    repo DomainGroupRepository
}
```

