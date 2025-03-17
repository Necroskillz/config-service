package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/necroskillz/config-service/model"
)

type VariationPropertyRepository struct {
	GormRepository[model.VariationProperty]
}

func NewVariationPropertyRepository(db *gorm.DB) *VariationPropertyRepository {
	return &VariationPropertyRepository{GormRepository[model.VariationProperty]{db: db}}
}

func (r *VariationPropertyRepository) GetAll(ctx context.Context) ([]model.VariationProperty, error) {
	var variationProperties []model.VariationProperty

	if err := r.getDb(ctx).Preload("ServiceTypes").Preload("Values").Find(&variationProperties).Error; err != nil {
		return nil, err
	}

	return variationProperties, nil
}
