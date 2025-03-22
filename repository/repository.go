package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type Repository[T any] interface {
	GetById(ctx context.Context, id uint) (*T, error)
	Create(ctx context.Context, user *T) error
	Update(ctx context.Context, user *T) error
	Delete(ctx context.Context, id uint) error
}

type GormRepository[T any] struct {
	db *gorm.DB
}

func NilIfNotFound[T any](entity *T, err error) (*T, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return entity, err
}

func (r *GormRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var entities []T

	err := r.getDb(ctx).Find(&entities).Error

	return entities, err
}

func (r *GormRepository[T]) GetByProperty(ctx context.Context, property string, value any, preload ...string) (*T, error) {
	var entity T

	db := r.getDb(ctx)

	for _, p := range preload {
		db = db.Preload(p)
	}

	result := db.Where(fmt.Sprintf("%s = ?", property), value).Limit(1).First(&entity)

	return NilIfNotFound(&entity, result.Error)
}

func (r *GormRepository[T]) GetById(ctx context.Context, id uint, preload ...string) (*T, error) {
	return r.GetByProperty(ctx, "id", id, preload...)
}

func (r *GormRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.getDb(ctx).Create(entity).Error
}

func (r *GormRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.getDb(ctx).Save(entity).Error
}

func (r *GormRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T

	return r.getDb(ctx).Delete(&entity, id).Error
}

func (r *GormRepository[T]) getDb(ctx context.Context) *gorm.DB {
	unitOfWork := ctx.Value(constants.UnitOfWorkKey)
	if unitOfWork == nil {
		return r.db.WithContext(ctx)
	}
	return unitOfWork.(*GormUnitOfWork).tx
}

func (r *VariationValueRepository) whereActiveOrInChangeset(db *gorm.DB, changesetID uint, changesetChangeIdName string, changesetChangeIdOldName string) *gorm.DB {
	isDeletedInChangeset := db.Table("changeset_changes csc").Where(fmt.Sprintf("csc.changeset_id = ? AND csc.type = ? AND csc.%s = o.id", changesetChangeIdOldName), changesetID, model.ChangesetChangeTypeDelete).Limit(1)
	isCreatedInChangeset := db.Table("changeset_changes csc").Where(fmt.Sprintf("csc.changeset_id = ? AND csc.type = ? AND csc.%s = o.id", changesetChangeIdName), changesetID, model.ChangesetChangeTypeCreate).Limit(1)

	return db.Where("(o.valid_from IS NOT NULL AND o.valid_to IS NULL AND NOT EXISTS (?)) OR (o.valid_from IS NULL AND o.valid_to IS NULL AND EXISTS (?))", isDeletedInChangeset, isCreatedInChangeset)
}
