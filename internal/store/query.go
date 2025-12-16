package store

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Query[T any] struct {
	db     *gorm.DB
	model  T
	limit  int
	offset int
}

func NewQuery[T any](db *gorm.DB) *Query[T] {
	var model T
	return &Query[T]{
		db:    db.Model(&model),
		model: model,
	}
}

func (q *Query[T]) Eq(field string, v any) *Query[T] {
	if v != nil {
		q.db = q.db.Where(fmt.Sprintf("%s = ?", field), v)
	}
	return q
}

func (q *Query[T]) Like(field string, v string) *Query[T] {
	if v != "" {
		q.db = q.db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+v+"%")
	}
	return q
}

func (q *Query[T]) In(field string, arr []any) *Query[T] {
	if len(arr) > 0 {
		q.db = q.db.Where(fmt.Sprintf("%s IN ?", field), arr)
	}
	return q
}

func (q *Query[T]) Between(field string, from any, to any) *Query[T] {
	if from != nil && to != nil {
		q.db = q.db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), from, to)
	}
	return q
}

func (q *Query[T]) Order(expr string) *Query[T] {
	if expr != "" {
		q.db = q.db.Order(expr)
	}
	return q
}

// OrderBy orders by field with direction (ASC or DESC)
func (q *Query[T]) OrderBy(field, direction string) *Query[T] {
	if field != "" {
		if direction == "" {
			direction = "ASC"
		}
		q.db = q.db.Order(fmt.Sprintf("%s %s", field, direction))
	}
	return q
}

func (q *Query[T]) Page(page, size int) *Query[T] {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	q.limit = size
	q.offset = (page - 1) * size
	return q
}

func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
	q.db = q.db.WithContext(ctx)
	return q
}

func (q *Query[T]) Count() (int64, error) {
	var count int64
	err := q.db.Count(&count).Error
	return count, err
}

func (q *Query[T]) Find(dest any) error {
	query := q.db
	if q.limit > 0 {
		query = query.Limit(q.limit)
	}
	if q.offset > 0 {
		query = query.Offset(q.offset)
	}
	return query.Find(dest).Error
}

func (q *Query[T]) First(dest any) error {
	return q.db.First(dest).Error
}
