package repository

import (
	"context"

	"gorm.io/gorm"
)

type baseRepository struct {
	db *gorm.DB
}

type BaseRepository interface {
	BeginTx(ctx context.Context) (*gorm.DB, error)
	CommitTx(ctx context.Context, tx *gorm.DB) (*gorm.DB, error)
	RollbackTx(ctx context.Context, tx *gorm.DB)
}

func NewBaseRepository(db *gorm.DB) BaseRepository {
	return &baseRepository{db: db}
}

// BeginTx starts a new transaction
func (r *baseRepository) BeginTx(ctx context.Context) (*gorm.DB, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

// CommitTx commits the transaction
func (r *baseRepository) CommitTx(ctx context.Context, tx *gorm.DB) (*gorm.DB, error) {
	err := tx.WithContext(ctx).Commit().Error
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// RollbackTx rolls back the transaction
func (r *baseRepository) RollbackTx(ctx context.Context, tx *gorm.DB) {
	tx.WithContext(ctx).Debug().Rollback()
}
