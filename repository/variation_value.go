package repository

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type VariationValueRepository struct {
	GormRepository[model.VariationValue]
}

func NewVariationValueRepository(db *gorm.DB) *VariationValueRepository {
	return &VariationValueRepository{GormRepository[model.VariationValue]{db: db}}
}

func (r *VariationValueRepository) GetActive(ctx context.Context, keyID uint, changesetID uint) ([]model.VariationValue, error) {
	var variationValues []model.VariationValue

	db := r.getDb(ctx)
	where := r.whereActiveOrInChangeset(db, changesetID, "new_variation_value_id", "old_variation_value_id")

	err := db.Table("variation_values o").
		Preload("VariationPropertyValues").
		Where("o.key_id = ?", keyID).
		Where(where).
		Find(&variationValues).Error

	return variationValues, err
}

func (r *VariationValueRepository) GetIDByVariation(ctx context.Context, keyID uint, variationIDs []uint, changesetID uint) (uint, error) {
	var id uint

	db := r.getDb(ctx)
	where := r.whereActiveOrInChangeset(db, changesetID, "new_variation_value_id", "old_variation_value_id")

	if len(variationIDs) > 0 {
		where = where.Where("vv.variation_property_value_id IN (?)", variationIDs)
	}

	err := db.
		Table("variation_values o").
		Select("o.id").
		Joins("LEFT JOIN variation_value_variation_property_values vv ON vv.variation_value_id = o.id").
		Where(where).
		Where("o.key_id = ?", keyID).
		Group("o.id").
		Having("COUNT(vv.variation_property_value_id) = ?", len(variationIDs)).
		Order("o.id DESC").
		Limit(1).
		Scan(&id).Error

	return id, err
}

func (r *VariationValueRepository) AddVariationPropertyValue(ctx context.Context, valueID uint, variationPropertyValueID uint) error {
	return r.getDb(ctx).Model(&model.VariationValue{ID: valueID}).Association("VariationPropertyValues").Append(&model.VariationPropertyValue{
		ID: variationPropertyValueID,
	})
}
