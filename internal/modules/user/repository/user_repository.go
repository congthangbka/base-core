package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/example/clean-architecture/internal/common"
	"github.com/example/clean-architecture/internal/entity"
	"github.com/example/clean-architecture/internal/store"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindAll(ctx context.Context, query *store.Query[entity.User]) ([]entity.User, error)
	Count(ctx context.Context, query *store.Query[entity.User]) (int64, error)
	FindAllWithFilters(ctx context.Context, name, email string, page, limit int) ([]entity.User, int64, error)
	WithTx(tx *gorm.DB) UserRepository
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) WithTx(tx *gorm.DB) UserRepository {
	return &userRepository{db: tx}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return common.WrapError(err, "failed to create user")
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).Model(&entity.User{}).Where(entity.Column.ID+" = ?", user.ID).Updates(user)
	if result.Error != nil {
		return common.WrapError(result.Error, "failed to update user")
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&entity.User{}, entity.Column.ID+" = ?", id)
	if result.Error != nil {
		return common.WrapError(result.Error, "failed to delete user")
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where(entity.Column.ID+" = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.WrapError(err, "failed to find user by id")
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where(entity.Column.Email+" = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.WrapError(err, "failed to find user by email")
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, query *store.Query[entity.User]) ([]entity.User, error) {
	var users []entity.User
	if err := query.WithContext(ctx).Find(&users); err != nil {
		return nil, common.WrapError(err, "failed to find users")
	}
	return users, nil
}

func (r *userRepository) Count(ctx context.Context, query *store.Query[entity.User]) (int64, error) {
	count, err := query.WithContext(ctx).Count()
	if err != nil {
		return 0, common.WrapError(err, "failed to count users")
	}
	return count, nil
}

func (r *userRepository) FindAllWithFilters(ctx context.Context, name, email string, page, limit int) ([]entity.User, int64, error) {
	// Build query using fluent query builder
	query := store.NewQuery[entity.User](r.db).WithContext(ctx)

	if name != "" {
		query = query.Like(entity.Column.Name, name)
	}
	if email != "" {
		query = query.Like(entity.Column.Email, email)
	}

	query = query.OrderBy(entity.Column.CreatedAt, entity.OrderDESC)
	query = query.Page(page, limit)

	// Get total count
	total, err := query.Count()
	if err != nil {
		return nil, 0, common.WrapError(err, "failed to count users")
	}

	// Get users
	var users []entity.User
	if err := query.Find(&users); err != nil {
		return nil, 0, common.WrapError(err, "failed to find users")
	}

	return users, total, nil
}
