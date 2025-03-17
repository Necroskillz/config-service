package repository

import (
	"context"
	"time"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type VariationValueRepository struct {
	GormRepository[model.VariationValue]
}

func NewVariationValueRepository(db *gorm.DB) *VariationValueRepository {
	return &VariationValueRepository{GormRepository[model.VariationValue]{db: db}}
}

func (r *VariationValueRepository) GetActive(ctx context.Context, keyID uint) ([]model.VariationValue, error) {
	var variationValues []model.VariationValue

	err := r.getDb(ctx).Preload("VariationPropertyValues").
		Where("key_id = ? AND valid_from <= ? AND (valid_to >= ? OR valid_to IS NULL)", keyID, time.Now(), time.Now()).
		Find(&variationValues).Error

	return variationValues, err
}
