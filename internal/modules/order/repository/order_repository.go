package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"llm-aggregator/internal/common"
	"llm-aggregator/internal/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	FindAllWithFilters(ctx context.Context, userID, productName string, status *int, page, limit int) ([]entity.Order, int64, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]entity.Order, int64, error)
	WithTx(tx *gorm.DB) OrderRepository
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return err
	}
	return nil
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	result := r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("id = ?", order.ID).
		Updates(map[string]interface{}{
			"product_name": order.ProductName,
			"quantity":     order.Quantity,
			"amount":       order.Amount,
			"status":       order.Status,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}

	return nil
}

func (r *orderRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Order{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}

	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAllWithFilters(ctx context.Context, userID, productName string, status *int, page, limit int) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if productName != "" {
		query = query.Where("product_name LIKE ?", "%"+productName+"%")
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{}).Where("user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) WithTx(tx *gorm.DB) OrderRepository {
	return &orderRepository{db: tx}
}

