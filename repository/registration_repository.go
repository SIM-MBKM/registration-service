package repository

import (
	"context"
	"errors"
	"registration-service/dto"
	"registration-service/entity"

	"gorm.io/gorm"
)

type registrationRepository struct {
	db             *gorm.DB
	baseRepository BaseRepository
}

type RegistrationRepository interface {
	Index(ctx context.Context, tx *gorm.DB, pagReq dto.PaginationRequest, filter dto.FilterDataRequest) ([]entity.Registration, int64, error)
	Create(ctx context.Context, registration entity.Registration, tx *gorm.DB) (entity.Registration, error)
	Update(ctx context.Context, id string, registration entity.Registration, tx *gorm.DB) error
	FindByID(ctx context.Context, id string, tx *gorm.DB) (entity.Registration, error)
	Destroy(ctx context.Context, id string, tx *gorm.DB) error
	FilterSubQuery(ctx context.Context, tx *gorm.DB, filter dto.FilterDataRequest) *gorm.DB
	FindTotal(ctx context.Context, tx *gorm.DB) (int64, error)
}

func NewRegistrationRepository(db *gorm.DB, baseRepository BaseRepository) RegistrationRepository {
	return &registrationRepository{db: db, baseRepository: baseRepository}
}

func (r *registrationRepository) FindTotal(ctx context.Context, tx *gorm.DB) (int64, error) {
	var total int64
	if tx == nil {
		tx = r.db
	}
	err := tx.WithContext(ctx).
		Model(&entity.Registration{}).
		Where("registrations.deleted_at IS NULL").
		Count(&total).Error

	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *registrationRepository) Index(ctx context.Context, tx *gorm.DB, pagReq dto.PaginationRequest, filter dto.FilterDataRequest) ([]entity.Registration, int64, error) {
	var registrations []entity.Registration
	if tx == nil {
		tx = r.db
	}

	subQuery := r.FilterSubQuery(ctx, tx, filter)

	err := subQuery.
		Preload("Documents").
		Offset(pagReq.Offset).
		Limit(pagReq.Limit).
		Find(&registrations).Error

	if err != nil {
		return nil, 0, err
	}

	total, err := r.FindTotal(ctx, tx)
	if err != nil {
		return nil, 0, err
	}

	return registrations, total, nil
}

func (r *registrationRepository) Create(ctx context.Context, registration entity.Registration, tx *gorm.DB) (entity.Registration, error) {
	tx, err := r.baseRepository.BeginTx(ctx)
	if err != nil {
		return entity.Registration{}, err
	}

	defer func() {
		if err != nil {
			r.baseRepository.RollbackTx(ctx, tx)
		}
	}()

	err = r.db.WithContext(ctx).
		Model(&entity.Registration{}).
		Create(&registration).Error
	if err != nil {
		return entity.Registration{}, err
	}

	r.baseRepository.CommitTx(ctx, tx)

	return registration, nil
}

func (r *registrationRepository) Update(ctx context.Context, id string, registration entity.Registration, tx *gorm.DB) error {
	tx, err := r.baseRepository.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			r.baseRepository.RollbackTx(ctx, tx)
		}
	}()

	data, err := r.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	if data.ID.String() == "" || data.ID.String() != id {
		return nil
	}

	err = r.db.WithContext(ctx).
		Model(&entity.Registration{}).
		Where("id = ?", id).
		Updates(&registration).Error

	if err != nil {
		return err
	}

	r.baseRepository.CommitTx(ctx, tx)
	return nil
}

func (r *registrationRepository) FindByID(ctx context.Context, id string, tx *gorm.DB) (entity.Registration, error) {
	var registration entity.Registration
	if tx == nil {
		tx = r.db
	}
	err := tx.WithContext(ctx).
		Model(&entity.Registration{}).
		Where("id = ?", id).
		First(&registration).Error

	if err != nil {
		return entity.Registration{}, err
	}

	return registration, nil
}

func (r *registrationRepository) Destroy(ctx context.Context, id string, tx *gorm.DB) error {
	tx, err := r.baseRepository.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			r.baseRepository.RollbackTx(ctx, tx)
		}
	}()

	data, err := r.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	if data.ID.String() == "" || data.ID.String() != id {
		return errors.New("data not found")
	}

	err = r.db.WithContext(ctx).
		Model(&entity.Registration{}).
		Where("id = ?", id).
		Delete(&entity.Registration{}).Error

	if err != nil {
		return err
	}

	r.baseRepository.CommitTx(ctx, tx)

	return nil
}

func (r *registrationRepository) FilterSubQuery(ctx context.Context, tx *gorm.DB, filter dto.FilterDataRequest) *gorm.DB {
	subQuery := tx.WithContext(ctx).
		Model(&entity.Registration{}).
		Where("registrations.deleted_at IS NULL")

	if len(filter.ActivityID) > 0 {
		subQuery = subQuery.Where("registrations.activity_id IN ?", filter.ActivityID)
	}

	if len(filter.UserID) > 0 {
		subQuery = subQuery.Where("registrations.user_id IN ?", filter.UserID)
	}

	if filter.AcademicAdvisor != "" {
		subQuery = subQuery.Where("registrations.academic_advisor = ?", filter.AcademicAdvisor)
	}

	if filter.ApprovalStatus {
		subQuery = subQuery.Where("registrations.approval_status = ?", filter.ApprovalStatus)
	}

	return subQuery
}
