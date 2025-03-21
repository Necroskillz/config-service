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

func (r *VariationValueRepository) GetIDByVariation(ctx context.Context, keyID uint, variationIDs []uint) (uint, error) {
	var id uint

	db := r.getDb(ctx)

	where := db.Where("vv.variation_property_value_id IN (?) AND v.key_id = ?", variationIDs, keyID)

	if len(variationIDs) == 0 {
		where = db.Where("v.key_id = ?", keyID)
	}

	err := r.getDb(ctx).
		Table("variation_values v").
		Select("v.id").
		Joins("LEFT JOIN variation_value_variation_property_values vv ON vv.variation_value_id = v.id").
		Where(where).
		Group("v.id").
		Having("COUNT(vv.variation_property_value_id) = ?", len(variationIDs)).
		Order("v.id DESC").
		Limit(1).
		Scan(&id).Error

	return id, err
}

func (r *VariationValueRepository) AddVariationPropertyValue(ctx context.Context, valueID uint, variationPropertyValueID uint) error {
	return r.getDb(ctx).Model(&model.VariationValue{ID: valueID}).Association("VariationPropertyValues").Append(&model.VariationPropertyValue{
		ID: variationPropertyValueID,
	})
}
