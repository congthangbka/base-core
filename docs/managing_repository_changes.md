# Managing Repository Method Changes - Best Practices

## The Concern

When a repository method is used in multiple places, changing it can affect many parts of the codebase. This is a valid concern about coupling and maintainability.

## Reality Check: Both Approaches Have This Issue

### Direct Store Usage
```go
// Used in 5 different services
store.ModifyRecord[entity.User](map[string]interface{}{"id": id}, updates)
store.ModifyRecord[entity.User](map[string]interface{}{"id": id}, updates)
store.ModifyRecord[entity.User](map[string]interface{}{"id": id}, updates)
// ... 2 more places

// When you change ModifyRecord signature or behavior:
// ❌ Must update ALL 5 places
// ❌ No type safety - compiler won't catch issues
// ❌ Hard to find all usages (no IDE support for generic functions)
```

### Repository Pattern
```go
// Used in 3 different service methods
s.repo.FindByID(ctx, id)  // In Update
s.repo.FindByID(ctx, id)  // In Delete  
s.repo.FindByID(ctx, id)  // In GetByID

// When you change FindByID signature:
// ✅ Interface enforces contract - compiler catches breaking changes
// ✅ IDE can find all usages easily
// ✅ Can add new method instead of changing existing one
```

**Verdict**: Repository pattern is **safer** because:
- Interface contracts prevent accidental breaking changes
- Compiler catches issues at compile time
- Better tooling support (find usages, refactoring)

## Best Practices to Minimize Impact

### 1. **Interface Stability - Don't Change Signatures**

**❌ Bad: Breaking Change**
```go
// Old signature
FindByID(ctx context.Context, id string) (*entity.User, error)

// New signature - BREAKS all callers
FindByID(ctx context.Context, id string, includeDeleted bool) (*entity.User, error)
```

**✅ Good: Add New Method**
```go
// Keep old method unchanged
FindByID(ctx context.Context, id string) (*entity.User, error)

// Add new method for new requirement
FindByIDWithDeleted(ctx context.Context, id string) (*entity.User, error)
```

### 2. **Composition Over Modification**

Instead of changing a method used everywhere, compose it:

```go
// Instead of changing FindByID to support soft deletes
func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    // Don't change this - it's used everywhere
    var user entity.User
    if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
        // ...
    }
    return &user, nil
}

// Add specific method for admin use case
func (r *userRepository) FindByIDIncludingDeleted(ctx context.Context, id string) (*entity.User, error) {
    var user entity.User
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
        // ...
    }
    return &user, nil
}
```

### 3. **Specific Methods for Specific Use Cases**

If different callers need different behavior, create specific methods:

```go
// Generic method - used in many places
FindByID(ctx context.Context, id string) (*entity.User, error)

// Specific method for update operations (needs optimistic locking)
FindByIDForUpdate(ctx context.Context, id string) (*entity.User, error)

// Specific method for admin operations (needs all fields)
FindByIDWithDetails(ctx context.Context, id string) (*entity.User, error)
```

### 4. **Backward Compatibility Strategy**

When you MUST change behavior, maintain backward compatibility:

```go
// Old method - keep for backward compatibility
func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    return r.FindByIDWithOptions(ctx, id, FindOptions{IncludeDeleted: false})
}

// New flexible method
func (r *userRepository) FindByIDWithOptions(ctx context.Context, id string, opts FindOptions) (*entity.User, error) {
    query := r.db.WithContext(ctx).Where("id = ?", id)
    if !opts.IncludeDeleted {
        query = query.Where("deleted_at IS NULL")
    }
    // ... rest of implementation
}
```

### 5. **Version Methods When Breaking Changes Are Unavoidable**

```go
// V1 - keep for existing code
FindByID(ctx context.Context, id string) (*entity.User, error)

// V2 - new signature, migrate gradually
FindByIDV2(ctx context.Context, id string, opts *FindOptions) (*entity.User, error)
```

## Real Example from Your Codebase

Looking at `user_service.go`, `FindByID` is used in 3 places:

```go
// Line 61: Update method
user, err := s.repo.FindByID(ctx, id)

// Line 101: Delete method  
_, err := s.repo.FindByID(ctx, id)

// Line 120: GetByID method
user, err := s.repo.FindByID(ctx, id)
```

### Scenario: Need to Add Soft Delete Support

**❌ Bad Approach: Change FindByID**
```go
// This breaks all 3 callers
func (r *userRepository) FindByID(ctx context.Context, id string, includeDeleted bool) (*entity.User, error) {
    // ...
}
```

**✅ Good Approach: Add New Method**
```go
// Keep existing method - no breaking change
func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    // Existing implementation - excludes deleted
    var user entity.User
    if err := r.db.WithContext(ctx).
        Where("id = ? AND deleted_at IS NULL", id).
        First(&user).Error; err != nil {
        // ...
    }
    return &user, nil
}

// Add new method for admin operations
func (r *userRepository) FindByIDIncludingDeleted(ctx context.Context, id string) (*entity.User, error) {
    var user entity.User
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
        // ...
    }
    return &user, nil
}
```

## Comparison: Repository vs Direct Store

| Aspect | Direct Store | Repository Pattern |
|--------|--------------|-------------------|
| **Breaking Changes** | ❌ No compiler protection | ✅ Interface enforces contract |
| **Find Usages** | ❌ Hard (generic functions) | ✅ Easy (IDE support) |
| **Refactoring** | ❌ Risky | ✅ Safe with interfaces |
| **Testing Impact** | ❌ Must update all mocks | ✅ Update interface once |
| **Type Safety** | ❌ Runtime errors | ✅ Compile-time errors |

## When to Split vs When to Share

### ✅ Share Method When:
- Behavior is **identical** across all use cases
- Requirements are **stable** (unlikely to change)
- Method is **simple** (single responsibility)

Example: `FindByID` - always finds by ID, simple and stable

### ✅ Split Method When:
- Different use cases need **different behavior**
- Some callers need **additional data** (e.g., relations)
- Some callers need **different query conditions**

Example:
```go
// Different behaviors needed
FindByID(ctx, id)                    // Basic lookup
FindByIDWithRelations(ctx, id)       // With joins
FindByIDForUpdate(ctx, id)           // With lock
FindByIDIncludingDeleted(ctx, id)    // Different filter
```

## Migration Strategy

If you need to change a widely-used method:

1. **Add new method** with new signature
2. **Keep old method** for backward compatibility
3. **Gradually migrate** callers to new method
4. **Deprecate old method** (add comment)
5. **Remove old method** after all migrations

```go
// Step 1: Add new method
func (r *userRepository) FindByIDV2(ctx context.Context, id string, opts *FindOptions) (*entity.User, error) {
    // New implementation
}

// Step 2: Keep old method, delegate to new
// Deprecated: Use FindByIDV2 instead
func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    return r.FindByIDV2(ctx, id, &FindOptions{IncludeDeleted: false})
}

// Step 3: Migrate callers one by one
// Step 4: Remove old method after migration
```

## Conclusion

**The concern is valid**, but:

1. **Repository pattern handles it better** than direct store usage
2. **Interface contracts** provide compile-time safety
3. **Best practices** minimize impact (add methods, don't change)
4. **Gradual migration** is possible with backward compatibility

**Recommendation**: Continue using Repository Pattern, but:
- Keep methods stable (don't change signatures)
- Add new methods for new requirements
- Use specific methods for specific use cases
- Plan migrations carefully when breaking changes are needed

