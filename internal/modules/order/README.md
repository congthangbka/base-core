# Order Module

## 📋 Tổng Quan

Module **Order** quản lý đơn hàng trong hệ thống. Module này cung cấp các chức năng CRUD cho order và có **inter-module communication** với User module để verify user và populate user information.

**Chức năng chính:**
- Tạo order mới với validation user exists và active
- Lấy thông tin order theo ID (kèm user info)
- Cập nhật order (product, quantity, amount, status)
- Xóa order
- Lấy danh sách order với pagination và filters
- Lấy orders theo user ID
- Transaction support cho atomic operations

---

## 🗄️ Database Table Structure

### Table: `orders`

| Column Name | Go Field | Type | Constraints | Description |
|------------|----------|------|-------------|-------------|
| `id` | `ID` | `varchar(36)` | PRIMARY KEY | UUID string |
| `user_id` | `UserID` | `varchar(36)` | NOT NULL, INDEX | Foreign key to users.id |
| `product_name` | `ProductName` | `varchar(255)` | NOT NULL | Tên sản phẩm |
| `quantity` | `Quantity` | `int` | NOT NULL, DEFAULT 1 | Số lượng |
| `amount` | `Amount` | `decimal(10,2)` | NOT NULL | Tổng tiền |
| `status` | `Status` | `int` | DEFAULT 1 | Trạng thái (1=pending, 2=completed, 3=cancelled) |
| `created_at` | `CreatedAt` | `timestamp` | AUTO | Thời gian tạo |
| `updated_at` | `UpdatedAt` | `timestamp` | AUTO | Thời gian cập nhật |

**Indexes:**
- Primary Key: `id`
- Index: `user_id` (for fast lookup by user)

**Entity Location:** `internal/entity/order.go`

**Status Constants:**
```go
OrderStatusPending   = 1  // Đang chờ xử lý
OrderStatusCompleted = 2  // Đã hoàn thành
OrderStatusCancelled = 3  // Đã hủy
```

---

## 📦 DTO Mapping

### Request DTOs

#### `CreateOrderRequest`
```go
{
    "userId":      string  (required)
    "productName": string  (required, 1-255 chars)
    "quantity":    int     (required, min=1)
    "amount":      float64 (required, min=0)
}
```

**Validation Rules:**
- `userId`: Required, must exist in users table, user must be active
- `productName`: Required, min=1, max=255
- `quantity`: Required, min=1
- `amount`: Required, min=0

#### `UpdateOrderRequest`
```go
{
    "productName": string   (optional, 1-255 chars)
    "quantity":    *int     (optional, min=1)
    "amount":      *float64 (optional, min=0)
    "status":      *int     (optional, 1 or 2 or 3)
}
```

**Validation Rules:**
- Tất cả fields đều optional
- `productName`: Nếu có thì min=1, max=255
- `quantity`: Nếu có thì min=1
- `amount`: Nếu có thì min=0
- `status`: Nếu có thì phải là 1, 2, hoặc 3

#### `OrderPagingRequest`
```go
{
    "page":        int     (optional, min=1, default=1)
    "limit":       int     (optional, min=1, max=100, default=10)
    "userId":      string  (optional, filter by user ID)
    "productName": string  (optional, filter by product name)
    "status":      *int    (optional, filter by status: 1, 2, or 3)
}
```

### Response DTOs

#### `OrderResponse`
```go
{
    "id":          string  (UUID)
    "userId":      string
    "userName":    string  (from User module via inter-module call)
    "userEmail":   string  (from User module via inter-module call)
    "productName": string
    "quantity":    int
    "amount":      float64
    "status":      int     (1=pending, 2=completed, 3=cancelled)
    "statusText":  string  ("pending", "completed", "cancelled")
    "createdAt":   string  (RFC3339 format)
    "updatedAt":   string  (RFC3339 format)
}
```

**Note:** `userName` và `userEmail` được populate từ User module thông qua inter-module communication. Nếu User module không available, các fields này sẽ là empty string.

#### `OrderPagingResponse`
```go
{
    "data":       []OrderResponse
    "page":       int
    "limit":      int
    "total":      int64
    "totalPages": int
}
```

### Entity → DTO Mapping

**Mapping Logic:** `service/order_service.go::toOrderResponse()`

| Entity Field | DTO Field | Transformation |
|--------------|-----------|----------------|
| `ID` | `id` | Direct mapping |
| `UserID` | `userId` | Direct mapping |
| `ProductName` | `productName` | Direct mapping |
| `Quantity` | `quantity` | Direct mapping |
| `Amount` | `amount` | Direct mapping |
| `Status` | `status` | Direct mapping |
| `Status` | `statusText` | `getStatusText(status)` → "pending"/"completed"/"cancelled" |
| `CreatedAt` | `createdAt` | `Format(time.RFC3339)` |
| `UpdatedAt` | `updatedAt` | `Format(time.RFC3339)` |
| - | `userName` | **Inter-module call:** `UserGetter.GetUserByID()` → `user.Name` |
| - | `userEmail` | **Inter-module call:** `UserGetter.GetUserByID()` → `user.Email` |

---

## 🔄 Logic Xử Lý Chi Tiết

### 1. Create Order (`POST /api/v1/orders`)

**Flow:**
```
Request → Handler → Validator → Service → User Module (verify) → Repository → Database
                                                      ↓
                                                Transaction Support
                                                      ↓
                                                Response (with user info)
```

**Chi tiết:**

1. **Handler** (`handler/order_handler.go::Create()`)
   - Bind JSON request → `CreateOrderRequest`
   - Validate request với `OrderValidator`
   - Gọi `service.Create()`

2. **Validator** (`validator/order_validator.go::ValidateCreateRequest()`)
   - Validate `userId`: required
   - Validate `productName`: required, min=1, max=255
   - Validate `quantity`: required, min=1
   - Validate `amount`: required, min=0

3. **Service** (`service/order_service.go::Create()`)
   - **Verify user exists and active (Inter-module call):**
     ```go
     user, err := s.getUserForValidation(ctx, req.UserID)
     if err != nil {
         return error: USER_NOT_FOUND
     }
     if user != nil && user.Status == 0 {
         return error: User is inactive
     }
     ```
     - Gọi `UserGetter.GetUserByID()` từ container
     - Check `user.Status == 0` (inactive)
   
   - **Create order entity:**
     ```go
     order := &entity.Order{
         ID:          uuid.New().String(),  // Generate UUID
         UserID:      req.UserID,
         ProductName: req.ProductName,
         Quantity:    req.Quantity,
         Amount:      req.Amount,
         Status:      entity.OrderStatusPending,  // Default: pending
     }
     ```
   
   - **Save to database (with transaction support):**
     ```go
     if s.db != nil {
         // Use transaction for atomic operation
         err := common.TransactionWithContext(ctx, s.db, func(tx *gorm.DB) error {
             txRepo := s.repo.WithTx(tx)
             return txRepo.Create(ctx, order)
         })
     } else {
         // Fallback to non-transactional create
         s.repo.Create(ctx, order)
     }
     ```
   
   - **Convert to response (populate user info):**
     ```go
     return s.toOrderResponse(ctx, order)
     ```

4. **Repository** (`repository/order_repository.go::Create()`)
   - Execute: `INSERT INTO orders (id, user_id, product_name, quantity, amount, status, created_at, updated_at) VALUES (...)`

**Inter-Module Communication:**
- **Call:** `container.UserGetter.GetUserByID(ctx, userID)`
- **Purpose:** Verify user exists và check status
- **Fallback:** Nếu UserGetter không available, chỉ verify existence (không check status)

**Error Codes:**
- `USER_NOT_FOUND`: User không tồn tại
- `INVALID`: User is inactive
- `VALIDATION_ERROR`: Validation failed
- `INTERNAL_ERROR`: Database error

---

### 2. Get Order By ID (`GET /api/v1/orders/:id`)

**Flow:**
```
Request → Handler → Service → Repository → Database
                                    ↓
                              Find By ID
                                    ↓
                              Populate User Info (Inter-module)
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/order_handler.go::GetByID()`)
   - Extract `id` từ path parameter
   - Gọi `service.GetByID()`

2. **Service** (`service/order_service.go::GetByID()`)
   - **Get order from repository:**
     ```go
     order, err := s.repo.FindByID(ctx, id)
     ```
   - **Convert to response (populate user info):**
     ```go
     return s.toOrderResponse(ctx, order)
     ```

3. **toOrderResponse()** - Populate user information:
   ```go
   // Get user info from User module (Inter-module call)
   if s.container.UserGetter != nil {
       user, err := s.container.UserGetter.GetUserByID(ctx, order.UserID)
       if err == nil && user != nil {
           response.UserName = user.Name
           response.UserEmail = user.Email
       }
       // Note: Silently ignore errors to avoid breaking response
       // if user service is temporarily unavailable
   }
   ```

**Inter-Module Communication:**
- **Call:** `container.UserGetter.GetUserByID(ctx, order.UserID)`
- **Purpose:** Populate `userName` và `userEmail` trong response
- **Error Handling:** Silently ignore errors (graceful degradation)

**Error Codes:**
- `NOT_FOUND`: Order không tồn tại
- `INTERNAL_ERROR`: Database error

---

### 3. Update Order (`PUT /api/v1/orders/:id`)

**Flow:**
```
Request → Handler → Validator → Service → Repository → Database
                                    ↓
                              Check Order Exists
                                    ↓
                              Update Fields
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/order_handler.go::Update()`)
   - Extract `id` từ path parameter
   - Bind JSON request → `UpdateOrderRequest`
   - Validate request
   - Gọi `service.Update()`

2. **Service** (`service/order_service.go::Update()`)
   - **Check order exists:**
     ```go
     order, err := s.repo.FindByID(ctx, id)
     ```
   - **Update fields (chỉ update fields có giá trị):**
     ```go
     if req.ProductName != "" {
         order.ProductName = req.ProductName
     }
     if req.Quantity != nil {
         order.Quantity = *req.Quantity
     }
     if req.Amount != nil {
         order.Amount = *req.Amount
     }
     if req.Status != nil {
         order.Status = *req.Status
     }
     ```
   - **Save changes:**
     ```go
     s.repo.Update(ctx, order)
     ```

3. **Repository** (`repository/order_repository.go::Update()`)
   - Execute: `UPDATE orders SET product_name=?, quantity=?, amount=?, status=?, updated_at=? WHERE id=?`
   - Return `ErrNotFound` nếu không tìm thấy

**Error Codes:**
- `NOT_FOUND`: Order không tồn tại
- `VALIDATION_ERROR`: Validation failed
- `INTERNAL_ERROR`: Database error

---

### 4. Delete Order (`DELETE /api/v1/orders/:id`)

**Flow:**
```
Request → Handler → Service → Repository → Database
                                    ↓
                              Check Order Exists
                                    ↓
                              Delete Order
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/order_handler.go::Delete()`)
   - Extract `id` từ path parameter
   - Gọi `service.Delete()`

2. **Service** (`service/order_service.go::Delete()`)
   - **Check order exists:**
     ```go
     _, err := s.repo.FindByID(ctx, id)
     ```
   - **Delete order:**
     ```go
     s.repo.Delete(ctx, id)
     ```

3. **Repository** (`repository/order_repository.go::Delete()`)
   - Execute: `DELETE FROM orders WHERE id = ?`
   - Return `ErrNotFound` nếu không tìm thấy

**Error Codes:**
- `NOT_FOUND`: Order không tồn tại
- `INTERNAL_ERROR`: Database error

---

### 5. Get All Orders (`GET /api/v1/orders`)

**Flow:**
```
Request → Handler → Validator → Service → Repository → Database
                                    ↓
                              Build Query with Filters
                                    ↓
                              Count Total
                                    ↓
                              Find Orders (with pagination)
                                    ↓
                              Populate User Info (Inter-module)
                                    ↓
                              Response
```

**Chi tiết:**

1. **Handler** (`handler/order_handler.go::GetAll()`)
   - Bind query parameters → `OrderPagingRequest`
   - Gọi `service.GetAll()`
   - Response với pagination helper

2. **Service** (`service/order_service.go::GetAll()`)
   - **Validate pagination:**
     ```go
     req.Page, req.Limit = common.ValidatePagination(
         req.Page, 
         req.Limit, 
         common.DefaultPaginationLimit  // Default: 10
     )
     ```
   - **Get orders with filters:**
     ```go
     orders, total, err := s.repo.FindAllWithFilters(
         ctx, 
         req.UserID,      // Filter by user ID
         req.ProductName, // Filter by product name (LIKE)
         req.Status,      // Filter by status
         req.Page, 
         req.Limit
     )
     ```
   - **Convert to responses (populate user info):**
     ```go
     orderResponses, err := s.convertOrdersToResponses(ctx, orders)
     ```
   - **Calculate total pages:**
     ```go
     TotalPages: common.CalculateTotalPages(total, req.Limit)
     ```

3. **Repository** (`repository/order_repository.go::FindAllWithFilters()`)
   - **Build query với filters:**
     ```go
     query := r.db.Model(&entity.Order{})
     if userID != "" {
         query = query.Where("user_id = ?", userID)
     }
     if productName != "" {
         query = query.Where("product_name LIKE ?", "%"+productName+"%")
     }
     if status != nil {
         query = query.Where("status = ?", *status)
     }
     query = query.Order("created_at DESC")
     query = query.Offset(offset).Limit(limit)
     ```
   - **Count total:**
     ```go
     query.Count(&total)
     ```
   - **Find orders:**
     ```go
     query.Find(&orders)
     ```

**SQL Example:**
```sql
-- Count
SELECT COUNT(*) FROM orders 
WHERE user_id = 'xxx' AND product_name LIKE '%laptop%' AND status = 1;

-- Find
SELECT * FROM orders 
WHERE user_id = 'xxx' AND product_name LIKE '%laptop%' AND status = 1
ORDER BY created_at DESC
LIMIT 10 OFFSET 0;
```

**Error Codes:**
- `VALIDATION_ERROR`: Validation failed
- `INTERNAL_ERROR`: Database error

---

### 6. Get Orders By User ID (`GET /api/v1/orders/user/:userId`)

**Flow:**
```
Request → Handler → Service → User Module (verify) → Repository → Database
                                                      ↓
                                                Populate User Info
                                                      ↓
                                                Response
```

**Chi tiết:**

1. **Handler** (`handler/order_handler.go::GetByUserID()`)
   - Extract `userId` từ path parameter
   - Extract `page`, `limit` từ query parameters
   - Gọi `service.GetByUserID()`

2. **Service** (`service/order_service.go::GetByUserID()`)
   - **Validate pagination:**
     ```go
     page, limit = common.ValidatePagination(page, limit, common.DefaultPaginationLimit)
     ```
   - **Verify user exists (Inter-module call):**
     ```go
     if err := s.verifyUserExists(ctx, userID); err != nil {
         return error: USER_NOT_FOUND
     }
     ```
     - Gọi `UserVerifier.VerifyUserExists()` từ container
     - Lightweight check (không fetch user data)
   
   - **Get orders by user ID:**
     ```go
     orders, total, err := s.repo.FindByUserID(ctx, userID, page, limit)
     ```
   
   - **Convert to responses (populate user info):**
     ```go
     orderResponses, err := s.convertOrdersToResponses(ctx, orders)
     ```

3. **Repository** (`repository/order_repository.go::FindByUserID()`)
   - Execute: `SELECT * FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

**Inter-Module Communication:**
- **Call:** `container.UserVerifier.VerifyUserExists(ctx, userID)`
- **Purpose:** Verify user exists trước khi query orders
- **Performance:** Lightweight check (không fetch user data)

**Error Codes:**
- `USER_NOT_FOUND`: User không tồn tại
- `INTERNAL_ERROR`: Database error

---

## 🌐 API Endpoints

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| `POST` | `/api/v1/orders` | `Create` | Tạo order mới |
| `GET` | `/api/v1/orders` | `GetAll` | Lấy danh sách order (pagination + filters) |
| `GET` | `/api/v1/orders/:id` | `GetByID` | Lấy order theo ID |
| `PUT` | `/api/v1/orders/:id` | `Update` | Cập nhật order |
| `DELETE` | `/api/v1/orders/:id` | `Delete` | Xóa order |
| `GET` | `/api/v1/orders/user/:userId` | `GetByUserID` | Lấy orders theo user ID |

**Route Registration:** `router.go::RegisterRoutes()`

---

## 🔗 Inter-Module Communication

Module Order **sử dụng** User module thông qua **type-safe interfaces** trong container.

### Interfaces Used

#### 1. `UserVerifier` (`interfaces.UserVerifier`)
**Purpose:** Verify user existence (lightweight, không fetch data)

**Method:**
```go
VerifyUserExists(ctx context.Context, userID string) error
```

**Used in:**
- `GetByUserID()` - Verify user exists trước khi query orders

**Implementation Location:** User module (`internal/modules/user/service/user_adapter.go`)

#### 2. `UserGetter` (`interfaces.UserGetter`)
**Purpose:** Get user information (fetch full data)

**Method:**
```go
GetUserByID(ctx context.Context, userID string) (*interfaces.UserInfo, error)
```

**Used in:**
- `Create()` - Verify user exists và check status
- `toOrderResponse()` - Populate `userName` và `userEmail` trong response

**Implementation Location:** User module (`internal/modules/user/service/user_adapter.go`)

### Usage Pattern

**1. Verify User Exists (Lightweight):**
```go
func (s *orderService) verifyUserExists(ctx context.Context, userID string) error {
    if s.container.UserVerifier == nil {
        return nil  // Graceful degradation
    }
    return s.container.UserVerifier.VerifyUserExists(ctx, userID)
}
```

**2. Get User for Validation:**
```go
func (s *orderService) getUserForValidation(ctx context.Context, userID string) (*interfaces.UserInfo, error) {
    if s.container.UserGetter == nil {
        // Fallback to verification only
        if err := s.verifyUserExists(ctx, userID); err != nil {
            return nil, err
        }
        return nil, nil
    }
    return s.container.UserGetter.GetUserByID(ctx, userID)
}
```

**3. Populate User Info in Response:**
```go
func (s *orderService) toOrderResponse(ctx context.Context, order *entity.Order) (*dto.OrderResponse, error) {
    // ... map order fields ...
    
    // Populate user info (graceful degradation)
    if s.container.UserGetter != nil {
        user, err := s.container.UserGetter.GetUserByID(ctx, order.UserID)
        if err == nil && user != nil {
            response.UserName = user.Name
            response.UserEmail = user.Email
        }
        // Silently ignore errors
    }
    
    return response, nil
}
```

### Container Dependency

**Location:** `service/order_service.go`

```go
type orderService struct {
    repo      repository.OrderRepository
    container *container.ModuleContainer  // For inter-module communication
    db        *gorm.DB                    // For transaction support
}
```

**Container Registration:** User module registers interfaces trong `router.go::RegisterRoutes()`

---

## 🔄 Transaction Support

Module Order hỗ trợ **database transactions** cho atomic operations.

### Implementation

**Service Constructor:**
```go
func NewOrderServiceWithDB(
    repo repository.OrderRepository, 
    container *container.ModuleContainer, 
    db *gorm.DB
) OrderService {
    return &orderService{
        repo:      repo,
        container: container,
        db:        db,  // Store db for transaction support
    }
}
```

**Transaction Usage (Create):**
```go
if s.db != nil {
    err := common.TransactionWithContext(ctx, s.db, func(tx *gorm.DB) error {
        txRepo := s.repo.WithTx(tx)  // Create repository with transaction
        return txRepo.Create(ctx, order)
    })
} else {
    // Fallback to non-transactional create
    s.repo.Create(ctx, order)
}
```

**Repository Transaction Support:**
```go
func (r *orderRepository) WithTx(tx *gorm.DB) OrderRepository {
    return &orderRepository{db: tx}  // Return new repository with transaction
}
```

**Transaction Helper:** `internal/common/transaction.go::TransactionWithContext()`

### Benefits

- **Atomic Operations:** Nếu có lỗi, tất cả changes sẽ rollback
- **Data Consistency:** Đảm bảo data consistency trong multi-step operations
- **Future-Proof:** Sẵn sàng cho các operations phức tạp hơn (multi-table updates)

---

## 📁 Module Structure

```
internal/modules/order/
├── README.md              # This file
├── router.go              # Route registration & dependency injection
├── dto/
│   └── order_dto.go       # Request/Response DTOs
├── handler/
│   └── order_handler.go   # HTTP handlers (Gin)
├── service/
│   ├── order_service.go   # Business logic + inter-module calls
│   └── order_adapter.go   # Inter-module adapter (for other modules)
├── repository/
│   └── order_repository.go # Database operations (GORM)
└── validator/
    └── order_validator.go # Request validation
```

---

## 🔧 Dependencies

### Internal Dependencies
- `internal/entity` - Order entity definition
- `internal/common` - Error handling, pagination helpers, transaction helpers
- `internal/container` - Module container for inter-module communication
- `internal/interfaces` - Inter-module interface definitions

### External Dependencies
- `gorm.io/gorm` - ORM for database operations
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/google/uuid` - UUID generation

### Inter-Module Dependencies
- **User Module:** Sử dụng `UserVerifier` và `UserGetter` interfaces

---

## 🎯 Key Design Patterns

1. **Clean Architecture / DDD:**
   - Separation of concerns: Handler → Service → Repository
   - Business logic trong Service layer
   - Data access trong Repository layer

2. **Adapter Pattern:**
   - `orderServiceAdapter` adapts `OrderService` to inter-module interfaces
   - Prevents circular dependencies

3. **Dependency Injection:**
   - Dependencies injected qua constructor
   - Container manages inter-module dependencies

4. **DRY (Don't Repeat Yourself):**
   - Common helpers: `HandleRepositoryError`, `ValidatePagination`, `CalculateTotalPages`
   - Reusable conversion methods: `toOrderResponse()`, `convertOrdersToResponses()`

5. **Graceful Degradation:**
   - Inter-module calls có fallback nếu service không available
   - Silently ignore errors khi populate user info (không break response)

---

## ⚠️ Error Handling

Module sử dụng centralized error handling từ `internal/common`:

**Error Types:**
- `ServiceError` - Structured error với code và message
- `ErrNotFound` - Entity không tồn tại
- `ErrInvalid` - Invalid input

**Error Codes:**
- `NOT_FOUND` - Order không tồn tại
- `USER_NOT_FOUND` - User không tồn tại (từ inter-module call)
- `INVALID` - User is inactive
- `VALIDATION_ERROR` - Validation failed
- `INTERNAL_ERROR` - Internal server error

**Error Handling Flow:**
```
Repository Error → HandleRepositoryError() → ServiceError → Handler → HTTP Response
Inter-Module Error → ServiceError → Handler → HTTP Response
```

---

## 🧪 Testing Considerations

**Test Cases to Cover:**

1. **Create Order:**
   - ✅ Valid order creation
   - ✅ User not found
   - ✅ User inactive
   - ✅ Invalid input (quantity < 1, amount < 0)
   - ✅ Transaction rollback on error

2. **Get Order:**
   - ✅ Order exists (with user info)
   - ✅ Order not found
   - ✅ User service unavailable (graceful degradation)

3. **Update Order:**
   - ✅ Update all fields
   - ✅ Update partial fields
   - ✅ Order not found

4. **Delete Order:**
   - ✅ Delete existing order
   - ✅ Delete non-existent order

5. **Get All Orders:**
   - ✅ Pagination
   - ✅ Filters (userId, productName, status)
   - ✅ Empty result

6. **Get Orders By User ID:**
   - ✅ User exists
   - ✅ User not found
   - ✅ Pagination

7. **Inter-Module Communication:**
   - ✅ UserVerifier available
   - ✅ UserVerifier unavailable (graceful degradation)
   - ✅ UserGetter available
   - ✅ UserGetter unavailable (graceful degradation)

---

## 📝 Notes

- **Status Values:**
  - `1` = Pending (default)
  - `2` = Completed
  - `3` = Cancelled

- **Pagination Defaults:**
  - Default page: `1`
  - Default limit: `10` (Order module specific)

- **Inter-Module Communication:**
  - Type-safe interfaces (compile-time checking)
  - Graceful degradation nếu service không available
  - Performance optimized: `UserVerifier` cho lightweight checks

- **Transaction Support:**
  - Hiện tại chỉ dùng trong `Create()` method
  - Repository hỗ trợ `WithTx()` cho future enhancements

- **User Info Population:**
  - `userName` và `userEmail` được populate từ User module
  - Nếu User module không available, các fields này sẽ là empty string
  - Không throw error nếu user lookup fails (graceful degradation)

---

## 🔄 Future Enhancements

Potential improvements:
- [ ] Order status workflow (pending → processing → completed)
- [ ] Order cancellation với reason
- [ ] Order history/audit log
- [ ] Order search với full-text search
- [ ] Order statistics/analytics
- [ ] Bulk operations (create multiple orders)
- [ ] Order export (CSV, Excel)
- [ ] Integration với payment gateway
- [ ] Order notifications

