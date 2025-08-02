package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// BaseModel represents the common fields for all database models
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Repository interface defines common repository operations
type Repository[T any] interface {
	// Basic CRUD operations
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error

	// Batch operations
	CreateMany(ctx context.Context, entities []T) error
	UpdateMany(ctx context.Context, entities []T) error
	DeleteMany(ctx context.Context, ids []uint) error

	// Query operations
	List(ctx context.Context, opts ListOptions) ([]T, error)
	Count(ctx context.Context, opts CountOptions) (int64, error)
	Exists(ctx context.Context, id uint) (bool, error)

	// Advanced operations
	FindWhere(ctx context.Context, conditions map[string]interface{}, opts ListOptions) ([]T, error)
	CountWhere(ctx context.Context, conditions map[string]interface{}) (int64, error)

	// Transaction support
	WithTx(tx *gorm.DB) Repository[T]
}

// ListOptions provides options for listing entities
type ListOptions struct {
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string // "asc" or "desc"
	Preloads   []string
	Conditions map[string]interface{}
}

// CountOptions provides options for counting entities
type CountOptions struct {
	Conditions map[string]interface{}
}

// BaseRepository provides a basic implementation of Repository interface
type BaseRepository[T any] struct {
	db    *gorm.DB
	model T
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	var model T
	return &BaseRepository[T]{
		db:    db,
		model: model,
	}
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID retrieves an entity by ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update updates an entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete soft deletes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

// CreateMany creates multiple entities in a batch
func (r *BaseRepository[T]) CreateMany(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, 100).Error
}

// UpdateMany updates multiple entities
func (r *BaseRepository[T]) UpdateMany(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Save(&entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteMany soft deletes multiple entities by IDs
func (r *BaseRepository[T]) DeleteMany(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, ids).Error
}

// List retrieves entities with options
func (r *BaseRepository[T]) List(ctx context.Context, opts ListOptions) ([]T, error) {
	var entities []T

	query := r.db.WithContext(ctx).Model(&r.model)

	// Apply conditions
	for key, value := range opts.Conditions {
		query = query.Where(key, value)
	}

	// Apply preloads
	for _, preload := range opts.Preloads {
		query = query.Preload(preload)
	}

	// Apply ordering
	if opts.OrderBy != "" {
		orderDir := "asc"
		if opts.OrderDir == "desc" {
			orderDir = "desc"
		}
		query = query.Order(opts.OrderBy + " " + orderDir)
	}

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	err := query.Find(&entities).Error
	return entities, err
}

// Count counts entities with options
func (r *BaseRepository[T]) Count(ctx context.Context, opts CountOptions) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&r.model)

	// Apply conditions
	for key, value := range opts.Conditions {
		query = query.Where(key, value)
	}

	err := query.Count(&count).Error
	return count, err
}

// Exists checks if an entity exists by ID
func (r *BaseRepository[T]) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&r.model).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindWhere finds entities with custom conditions
func (r *BaseRepository[T]) FindWhere(
	ctx context.Context,
	conditions map[string]interface{},
	opts ListOptions,
) ([]T, error) {
	// Merge conditions
	if opts.Conditions == nil {
		opts.Conditions = conditions
	} else {
		for key, value := range conditions {
			opts.Conditions[key] = value
		}
	}

	return r.List(ctx, opts)
}

// CountWhere counts entities with custom conditions
func (r *BaseRepository[T]) CountWhere(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return r.Count(ctx, CountOptions{Conditions: conditions})
}

// WithTx returns a new repository instance with the given transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) Repository[T] {
	return &BaseRepository[T]{
		db:    tx,
		model: r.model,
	}
}

// GetTableName returns the table name for the entity
func (r *BaseRepository[T]) GetTableName() string {
	stmt := &gorm.Statement{DB: r.db}
	if err := stmt.Parse(&r.model); err != nil {
		// Log error but return a fallback - this shouldn't fail in normal circumstances
		return "unknown_table"
	}
	return stmt.Schema.Table
}

// SoftDelete permanently deletes an entity (bypass soft delete)
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Unscoped().Delete(&entity, id).Error
}

// Restore restores a soft-deleted entity
func (r *BaseRepository[T]) Restore(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Unscoped().Model(&entity).Where("id = ?", id).Update("deleted_at", nil).Error
}

// BatchProcessor provides batch processing capabilities
type BatchProcessor[T any] struct {
	repo      Repository[T]
	batchSize int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor[T any](repo Repository[T], batchSize int) *BatchProcessor[T] {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &BatchProcessor[T]{
		repo:      repo,
		batchSize: batchSize,
	}
}

// ProcessInBatches processes entities in batches
func (bp *BatchProcessor[T]) ProcessInBatches(
	ctx context.Context,
	entities []T,
	processor func(batch []T) error,
) error {
	for i := 0; i < len(entities); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(entities) {
			end = len(entities)
		}

		batch := entities[i:end]
		if err := processor(batch); err != nil {
			return err
		}
	}
	return nil
}

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder[T any] struct {
	db         *gorm.DB
	conditions []func(*gorm.DB) *gorm.DB
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder[T any](db *gorm.DB) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		db:         db,
		conditions: make([]func(*gorm.DB) *gorm.DB, 0),
	}
}

// Where adds a where condition
func (qb *QueryBuilder[T]) Where(query interface{}, args ...interface{}) *QueryBuilder[T] {
	qb.conditions = append(qb.conditions, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return qb
}

// Order adds an order clause
func (qb *QueryBuilder[T]) Order(value interface{}) *QueryBuilder[T] {
	qb.conditions = append(qb.conditions, func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
	})
	return qb
}

// Limit adds a limit clause
func (qb *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	qb.conditions = append(qb.conditions, func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	})
	return qb
}

// Offset adds an offset clause
func (qb *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	qb.conditions = append(qb.conditions, func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	})
	return qb
}

// Preload adds a preload clause
func (qb *QueryBuilder[T]) Preload(query string, args ...interface{}) *QueryBuilder[T] {
	qb.conditions = append(qb.conditions, func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args...)
	})
	return qb
}

// Build builds the final query
func (qb *QueryBuilder[T]) Build() *gorm.DB {
	query := qb.db
	for _, condition := range qb.conditions {
		query = condition(query)
	}
	return query
}

// Find executes the query and returns results
func (qb *QueryBuilder[T]) Find(ctx context.Context) ([]T, error) {
	var results []T
	var model T
	err := qb.Build().WithContext(ctx).Model(&model).Find(&results).Error
	return results, err
}

// First executes the query and returns the first result
func (qb *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	var result T
	err := qb.Build().WithContext(ctx).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Count executes the query and returns the count
func (qb *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var model T
	err := qb.Build().WithContext(ctx).Model(&model).Count(&count).Error
	return count, err
}
