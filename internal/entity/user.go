package entity

import (
	"time"
)

type User struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Status    int       `gorm:"type:int;default:1" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// Column contains all database column names for User entity
var Column = struct {
	ID        string
	Name      string
	Email     string
	Status    string
	CreatedAt string
	UpdatedAt string
}{
	ID:        "id",
	Name:      "name",
	Email:     "email",
	Status:    "status",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// UserTableName is the table name for User entity
const UserTableName = "users"

// Order directions
const (
	OrderASC  = "ASC"
	OrderDESC = "DESC"
)

func (User) TableName() string {
	return UserTableName
}
