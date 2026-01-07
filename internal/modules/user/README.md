# User Module

## 📋 Tổng Quan

Module **User** quản lý thông tin người dùng trong hệ thống. Module này cung cấp các chức năng CRUD (Create, Read, Update, Delete) cho user và hỗ trợ tìm kiếm, phân trang.

**Chức năng chính:**
- Tạo user mới với validation email unique
- Lấy thông tin user theo ID
- Cập nhật thông tin user
- Xóa user
- Lấy danh sách user với pagination và filters (name, email)
- Cung cấp inter-module interfaces cho các module khác sử dụng

---

## 🗄️ Database Table Structure

### Table: `users`

| Column Name | Go Field | Type | Constraints | Description |
|------------|----------|------|-------------|-------------|
| `id` | `ID` | `varchar(36)` | PRIMARY KEY | UUID string |
| `name` | `Name` | `varchar(255)` | NOT NULL | Tên user |
| `email` | `Email` | `varchar(255)` | UNIQUE, NOT NULL | Email user (unique) |
| `status` | `Status` | `int` | DEFAULT 1 | Trạng thái (0=inactive, 1=active) |
| `created_at` | `CreatedAt` | `timestamp` | AUTO | Thời gian tạo |
| `updated_at` | `UpdatedAt` | `timestamp` | AUTO | Thời gian cập nhật |

**Indexes:**
- Primary Key: `id`
- Unique Index: `email`

**Entity Location:** `internal/entity/user.go`

---

## 📦 DTO Mapping

### Request DTOs

#### `CreateUserRequest`
```go
{
    "name":  string (required, 1-255 chars)
    "email": string (required, valid email format)
}
```

**Validation Rules:**
- `name`: Required, min=1, max=255
- `email`: Required, valid email format, unique in database

#### `UpdateUserRequest`
```go
{
    "name":   string (optional, 1-255 chars)
    "email":  string (optional, valid email format)
    "status": int    (optional, 0 or 1)
}
```

**Validation Rules:**
- Tất cả fields đều optional
- `name`: Nếu có thì min=1, max=255
- `email`: Nếu có thì valid email format, unique
- `status`: Nếu có thì phải là 0 hoặc 1

#### `PagingRequest`
```go
{
    "page":  int    (optional, min=1, default=1)
    "limit": int    (optional, min=1, max=100, default=20)
    "name":  string (optional, filter by name)
    "email": string (optional, filter by email)
}
```

### Response DTOs

#### `UserResponse`
```go
{
    "id":        string (UUID)
    "name":      string
    "email":     string
    "status":    int    (0=inactive, 1=active)
    "createdAt": string (RFC3339 format)
    "updatedAt": string (RFC3339 format)
}
```

#### `UserPagingResponse`
```go
{
    "data":       []UserResponse
    "page":       int
    "limit":      int
    "total":      int64
    "totalPages": int
}
```

### Entity → DTO Mapping

**Mapping Logic:** `service/user_service.go::toUserResponse()`

| Entity Field | DTO Field | Transformation |
|--------------|-----------|-----------------|
| `ID` | `id` | Direct mapping |
| `Name` | `name` | Direct mapping |
| `Email` | `email` | Direct mapping |
| `Status` | `status` | Direct mapping |
| `CreatedAt` | `createdAt` | `Format(time.RFC3339)` |
| `UpdatedAt` | `updatedAt` | `Format(time.RFC3339)` |

---

## 🔄 Logic Xử Lý Chi Tiết

### 1. Create User (`POST /api/v1/users`)

**Flow:**
```
Request → Handler → Validator → Service → Repository → Database
                                    ↓
                              Check Email Exists
                                    ↓
                              Create User (Status=1)
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/user_handler.go::Create()`)
   - Bind JSON request → `CreateUserRequest`
   - Validate request với `UserValidator`
   - Gọi `service.Create()`

2. **Validator** (`validator/user_validator.go::ValidateCreate()`)
   - Validate `name`: required, min=1, max=255
   - Validate `email`: required, valid email format

3. **Service** (`service/user_service.go::Create()`)
   - **Check email uniqueness:**
     ```go
     existingUser, err := s.repo.FindByEmail(ctx, req.Email)
     if existingUser != nil {
         return error: EMAIL_EXISTS
     }
     ```
   - **Create user entity:**
     ```go
     user := &entity.User{
         ID:     uuid.New().String(),  // Generate UUID
         Name:   req.Name,
         Email:  req.Email,
         Status: 1,  // Default: active
     }
     ```
   - **Save to database:**
     ```go
     s.repo.Create(ctx, user)
     ```
   - **Convert to response:**
     ```go
     return s.toUserResponse(user)
     ```

4. **Repository** (`repository/user_repository.go::Create()`)
   - Execute: `INSERT INTO users (id, name, email, status, created_at, updated_at) VALUES (...)`

**Error Codes:**
- `EMAIL_EXISTS`: Email đã tồn tại
- `VALIDATION_ERROR`: Validation failed
- `INTERNAL_ERROR`: Database error

---

### 2. Get User By ID (`GET /api/v1/users/:id`)

**Flow:**
```
Request → Handler → Service → Repository → Database
                                    ↓
                              Find By ID
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/user_handler.go::GetByID()`)
   - Extract `id` từ path parameter
   - Gọi `service.GetByID()`

2. **Service** (`service/user_service.go::GetByID()`)
   - **Get user from repository:**
     ```go
     user, err := s.repo.FindByID(ctx, id)
     ```
   - **Convert to response:**
     ```go
     return s.toUserResponse(user)
     ```

3. **Repository** (`repository/user_repository.go::FindByID()`)
   - Execute: `SELECT * FROM users WHERE id = ?`
   - Return `ErrNotFound` nếu không tìm thấy

**Error Codes:**
- `USER_NOT_FOUND`: User không tồn tại
- `INTERNAL_ERROR`: Database error

---

### 3. Update User (`PUT /api/v1/users/:id`)

**Flow:**
```
Request → Handler → Validator → Service → Repository → Database
                                    ↓
                              Check User Exists
                                    ↓
                              Check Email Uniqueness (if email changed)
                                    ↓
                              Update Fields
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/user_handler.go::Update()`)
   - Extract `id` từ path parameter
   - Bind JSON request → `UpdateUserRequest`
   - Validate request
   - Gọi `service.Update()`

2. **Service** (`service/user_service.go::Update()`)
   - **Check user exists:**
     ```go
     user, err := s.repo.FindByID(ctx, id)
     ```
   - **Check email uniqueness (nếu email thay đổi):**
     ```go
     if req.Email != "" && req.Email != user.Email {
         existingUser, err := s.repo.FindByEmail(ctx, req.Email)
         if existingUser != nil {
             return error: EMAIL_EXISTS
         }
         user.Email = req.Email
     }
     ```
   - **Update fields (chỉ update fields có giá trị):**
     ```go
     if req.Name != "" {
         user.Name = req.Name
     }
     if req.Status != nil {
         user.Status = *req.Status
     }
     ```
   - **Save changes:**
     ```go
     s.repo.Update(ctx, user)
     ```

3. **Repository** (`repository/user_repository.go::Update()`)
   - Execute: `UPDATE users SET name=?, email=?, status=?, updated_at=? WHERE id=?`
   - Return `ErrNotFound` nếu không tìm thấy

**Error Codes:**
- `USER_NOT_FOUND`: User không tồn tại
- `EMAIL_EXISTS`: Email đã tồn tại (nếu email thay đổi)
- `VALIDATION_ERROR`: Validation failed
- `INTERNAL_ERROR`: Database error

---

### 4. Delete User (`DELETE /api/v1/users/:id`)

**Flow:**
```
Request → Handler → Service → Repository → Database
                                    ↓
                              Check User Exists
                                    ↓
                              Delete User
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/user_handler.go::Delete()`)
   - Extract `id` từ path parameter
   - Gọi `service.Delete()`

2. **Service** (`service/user_service.go::Delete()`)
   - **Check user exists:**
     ```go
     _, err := s.repo.FindByID(ctx, id)
     ```
   - **Delete user:**
     ```go
     s.repo.Delete(ctx, id)
     ```

3. **Repository** (`repository/user_repository.go::Delete()`)
   - Execute: `DELETE FROM users WHERE id = ?`
   - Return `ErrNotFound` nếu không tìm thấy

**Error Codes:**
- `USER_NOT_FOUND`: User không tồn tại
- `INTERNAL_ERROR`: Database error

---

### 5. Get All Users (`GET /api/v1/users`)

**Flow:**
```
Request → Handler → Validator → Service → Repository → Database
                                    ↓
                              Build Query with Filters
                                    ↓
                              Count Total
                                    ↓
                              Find Users (with pagination)
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/user_handler.go::GetAll()`)
   - Bind query parameters → `PagingRequest`
   - Validate request
   - Gọi `service.GetAll()`
   - Response với pagination helper

2. **Service** (`service/user_service.go::GetAll()`)
   - **Validate pagination:**
     ```go
     req.Page, req.Limit = common.ValidatePagination(
         req.Page, 
         req.Limit, 
         common.DefaultPaginationLimitUser  // Default: 20
     )
     ```
   - **Get users with filters:**
     ```go
     users, total, err := s.repo.FindAllWithFilters(
         ctx, 
         req.Name,   // Filter by name (LIKE)
         req.Email,  // Filter by email (LIKE)
         req.Page, 
         req.Limit
     )
     ```
   - **Convert to responses:**
     ```go
     userResponses := s.convertUsersToResponses(users)
     ```
   - **Calculate total pages:**
     ```go
     TotalPages: common.CalculateTotalPages(total, req.Limit)
     ```

3. **Repository** (`repository/user_repository.go::FindAllWithFilters()`)
   - **Build query với filters:**
     ```go
     query := store.NewQuery[entity.User](r.db)
     if name != "" {
         query = query.Like(entity.Column.Name, name)  // WHERE name LIKE '%name%'
     }
     if email != "" {
         query = query.Like(entity.Column.Email, email)  // WHERE email LIKE '%email%'
     }
     query = query.OrderBy(entity.Column.CreatedAt, entity.OrderDESC)
     query = query.Page(page, limit)
     ```
   - **Count total:**
     ```go
     total, err := query.Count()
     ```
   - **Find users:**
     ```go
     query.Find(&users)
     ```

**SQL Example:**
```sql
-- Count
SELECT COUNT(*) FROM users 
WHERE name LIKE '%john%' AND email LIKE '%@gmail.com%';

-- Find
SELECT * FROM users 
WHERE name LIKE '%john%' AND email LIKE '%@gmail.com%'
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;
```

**Error Codes:**
- `VALIDATION_ERROR`: Validation failed
- `INTERNAL_ERROR`: Database error

---

## 🌐 API Endpoints

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| `POST` | `/api/v1/users` | `Create` | Tạo user mới |
| `GET` | `/api/v1/users` | `GetAll` | Lấy danh sách user (pagination + filters) |
| `GET` | `/api/v1/users/:id` | `GetByID` | Lấy user theo ID |
| `PUT` | `/api/v1/users/:id` | `Update` | Cập nhật user |
| `DELETE` | `/api/v1/users/:id` | `Delete` | Xóa user |

**Route Registration:** `router.go::RegisterRoutes()`

---

## 🔗 Inter-Module Communication

Module User cung cấp interfaces cho các module khác sử dụng thông qua **Adapter Pattern**.

### Interfaces Provided

#### 1. `UserVerifier` (`interfaces.UserVerifier`)
**Purpose:** Verify user existence (lightweight, không fetch data)

**Method:**
```go
VerifyUserExists(ctx context.Context, userID string) error
```

**Implementation:** `service/user_adapter.go::VerifyUserExists()`
- Gọi `service.GetByID()` để verify
- Return error nếu user không tồn tại

**Used by:** Order module (verify user exists trước khi tạo order)

#### 2. `UserGetter` (`interfaces.UserGetter`)
**Purpose:** Get user information (fetch full data)

**Method:**
```go
GetUserByID(ctx context.Context, userID string) (*interfaces.UserInfo, error)
```

**Implementation:** `service/user_adapter.go::GetUserByID()`
- Gọi `service.GetByID()`
- Convert `UserResponse` → `interfaces.UserInfo`

**Used by:** Order module (populate user name/email trong order response)

### Registration

**Location:** `router.go::RegisterRoutes()`

```go
// Create adapter
userAdapter := service.NewUserServiceAdapter(userService)

// Register in container
container.SetUserVerifier(userAdapter)  // For verification
container.SetUserGetter(userAdapter)    // For getting user data
```

**Container Location:** `internal/container/container.go`

---

## 📁 Module Structure

```
internal/modules/user/
├── README.md              # This file
├── router.go              # Route registration & dependency injection
├── dto/
│   └── user_dto.go        # Request/Response DTOs
├── handler/
│   └── user_handler.go    # HTTP handlers (Gin)
├── service/
│   ├── user_service.go           # Business logic
│   ├── user_adapter.go           # Inter-module adapter
│   └── user_service_metrics.go   # Prometheus metrics
├── repository/
│   └── user_repository.go # Database operations (GORM)
└── validator/
    └── user_validator.go  # Request validation
```

---

## 🔧 Dependencies

### Internal Dependencies
- `internal/entity` - User entity definition
- `internal/common` - Error handling, pagination helpers
- `internal/container` - Module container for inter-module communication
- `internal/interfaces` - Inter-module interface definitions
- `internal/store` - Query builder

### External Dependencies
- `gorm.io/gorm` - ORM for database operations
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/google/uuid` - UUID generation

---

## 🎯 Key Design Patterns

1. **Clean Architecture / DDD:**
   - Separation of concerns: Handler → Service → Repository
   - Business logic trong Service layer
   - Data access trong Repository layer

2. **Adapter Pattern:**
   - `userServiceAdapter` adapts `UserService` to inter-module interfaces
   - Prevents circular dependencies

3. **Dependency Injection:**
   - Dependencies injected qua constructor
   - Container manages inter-module dependencies

4. **DRY (Don't Repeat Yourself):**
   - Common helpers: `HandleRepositoryError`, `ValidatePagination`, `CalculateTotalPages`
   - Reusable conversion methods: `toUserResponse()`, `convertUsersToResponses()`

---

## 📊 Metrics & Observability

**Location:** `service/user_service_metrics.go`

Module sử dụng **Prometheus metrics** để track:
- Request count per endpoint
- Request duration
- Error count

**Metrics:**
- `user_service_requests_total` - Total requests
- `user_service_request_duration_seconds` - Request duration
- `user_service_errors_total` - Total errors

**Instrumentation:** `NewInstrumentedUserService()` wraps base service với metrics

---

## ⚠️ Error Handling

Module sử dụng centralized error handling từ `internal/common`:

**Error Types:**
- `ServiceError` - Structured error với code và message
- `ErrNotFound` - Entity không tồn tại
- `ErrInvalid` - Invalid input

**Error Codes:**
- `USER_NOT_FOUND` - User không tồn tại
- `EMAIL_EXISTS` - Email đã tồn tại
- `VALIDATION_ERROR` - Validation failed
- `INTERNAL_ERROR` - Internal server error

**Error Handling Flow:**
```
Repository Error → HandleRepositoryError() → ServiceError → Handler → HTTP Response
```

---

## 🧪 Testing Considerations

**Test Cases to Cover:**

1. **Create User:**
   - ✅ Valid user creation
   - ✅ Duplicate email
   - ✅ Invalid email format
   - ✅ Missing required fields

2. **Get User:**
   - ✅ User exists
   - ✅ User not found

3. **Update User:**
   - ✅ Update all fields
   - ✅ Update partial fields
   - ✅ Email uniqueness check
   - ✅ User not found

4. **Delete User:**
   - ✅ Delete existing user
   - ✅ Delete non-existent user

5. **Get All Users:**
   - ✅ Pagination
   - ✅ Filters (name, email)
   - ✅ Empty result

---

## 📝 Notes

- **Status Values:**
  - `0` = Inactive
  - `1` = Active (default)

- **Pagination Defaults:**
  - Default page: `1`
  - Default limit: `20` (User module specific)

- **Email Uniqueness:**
  - Enforced at database level (UNIQUE constraint)
  - Also checked in service layer before create/update

- **Transaction Support:**
  - Repository supports `WithTx()` for transactions
  - Currently not used in User module (no multi-table operations)

---

## 🔄 Future Enhancements

Potential improvements:
- [ ] Soft delete (thay vì hard delete)
- [ ] User roles/permissions
- [ ] Password management
- [ ] Email verification
- [ ] User search với full-text search
- [ ] Caching layer cho frequently accessed users

