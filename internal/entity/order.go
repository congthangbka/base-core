package entity

import (
	"time"
)

type Order struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;index" json:"userId"`
	ProductName string    `gorm:"type:varchar(255);not null" json:"productName"`
	Quantity    int       `gorm:"type:int;not null;default:1" json:"quantity"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status      int       `gorm:"type:int;default:1" json:"status"` // 1: pending, 2: completed, 3: cancelled
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// OrderColumn contains all database column names for Order entity
var OrderColumn = struct {
	ID          string
	UserID      string
	ProductName string
	Quantity    string
	Amount      string
	Status      string
	CreatedAt   string
	UpdatedAt   string
}{
	ID:          "id",
	UserID:      "user_id",
	ProductName: "product_name",
	Quantity:    "quantity",
	Amount:      "amount",
	Status:      "status",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// OrderTableName is the table name for Order entity
const OrderTableName = "orders"

// Order status constants
const (
	OrderStatusPending   = 1
	OrderStatusCompleted = 2
	OrderStatusCancelled = 3
)

func (Order) TableName() string {
	return OrderTableName
}

