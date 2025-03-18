package repository

import (
	"context"
	"registration-service/dto"
	"registration-service/entity"
	"time"

	"gorm.io/gorm"
)

type documentRepository struct {
	db             *gorm.DB
	baseRepository BaseRepository
}

type DocumentRepository interface {
	Index(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]entity.Document, int64, error)
	Create(ctx context.Context, document entity.Document, tx *gorm.DB) (entity.Document, error)
	Update(ctx context.Context, id string, document entity.Document, tx *gorm.DB) error
	FindByID(ctx context.Context, id string, tx *gorm.DB) (entity.Document, error)
	DeleteByID(ctx context.Context, id string, tx *gorm.DB) error
	FindTotal(ctx context.Context, tx *gorm.DB) (int64, error)
	GetAll() ([]entity.Document, error)
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{
		db:             db,
		baseRepository: NewBaseRepository(db),
	}
}

func (r *documentRepository) GetAll() ([]entity.Document, error) {
	var documents []entity.Document
	result := r.db.Find(&documents)

	return documents, result.Error
}

func (r *documentRepository) FindTotal(ctx context.Context, tx *gorm.DB) (int64, error) {
	var total int64
	if tx == nil {
		tx = r.db
	}
	err := tx.WithContext(ctx).
		Model(&entity.Document{}).
		Where("documents.deleted_at IS NULL").
		Count(&total).Error

	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *documentRepository) Index(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]entity.Document, int64, error) {
	var documents []entity.Document
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).
		Model(&entity.Document{}).
		Offset(pagReq.Offset).
		Limit(pagReq.Limit).
		First(&documents).Error

	if err != nil {
		return []entity.Document{}, 0, err
	}

	total, err := r.FindTotal(ctx, tx)
	if err != nil {
		return []entity.Document{}, 0, err
	}

	return documents, total, nil
}

func (r *documentRepository) FindByID(ctx context.Context, id string, tx *gorm.DB) (entity.Document, error) {
	var document entity.Document

	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).
		Model(&entity.Document{}).
		Where("id = ?", id).
		Find(&document).Error

	if err != nil {
		return entity.Document{}, err
	}

	return document, nil
}

func (r *documentRepository) Create(ctx context.Context, document entity.Document, tx *gorm.DB) (entity.Document, error) {

	tx, err := r.baseRepository.BeginTx(ctx)
	if err != nil {
		return entity.Document{}, err
	}

	defer func() {
		if err != nil {
			r.baseRepository.RollbackTx(ctx, tx)
		}
	}()

	err = tx.WithContext(ctx).Create(&document).Error
	if err != nil {
		return entity.Document{}, err
	}

	_, err = r.baseRepository.CommitTx(ctx, tx)
	if err != nil {
		return entity.Document{}, err
	}

	return document, nil
}

func (r *documentRepository) Update(ctx context.Context, id string, document entity.Document, tx *gorm.DB) error {

	var documentEntity entity.Document

	err := r.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", id).Find(&documentEntity).Error
	if err != nil {
		return err
	}

	tx, err = r.baseRepository.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", id).Updates(document).Error
	if tx != nil {
		return err
	}

	defer func() {
		if err != nil {
			r.baseRepository.RollbackTx(ctx, tx)
		}
	}()

	_, err = r.baseRepository.CommitTx(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

func (r *documentRepository) DeleteByID(ctx context.Context, id string, tx *gorm.DB) error {
	var documentEntity entity.Document
	tx, err := r.baseRepository.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", id).Find(&documentEntity).Error
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", id).UpdateColumn("deleted_at", time.Now()).Error

	if err != nil {
		return err
	}
	_, err = r.baseRepository.CommitTx(ctx, tx)

	defer func() {
		if err != nil {
			r.baseRepository.RollbackTx(ctx, tx)
		}
	}()

	return nil
}
