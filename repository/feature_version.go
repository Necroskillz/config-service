package repository

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type FeatureVersionRepository struct {
	GormRepository[model.FeatureVersion]
}

func NewFeatureVersionRepository(db *gorm.DB) *FeatureVersionRepository {
	return &FeatureVersionRepository{GormRepository[model.FeatureVersion]{db: db}}
}

func (r *FeatureVersionRepository) GetByFeatureIDForServiceVersion(ctx context.Context, featureID uint, serviceVersionID uint) ([]model.FeatureVersion, error) {
	var result []model.FeatureVersion

	subQuery := r.getDb(ctx).Table("feature_version_service_versions").Select("1").Limit(1).
		Where("service_version_id = ? AND feature_version_id = ? AND valid_from IS NOT NULL", serviceVersionID, featureID)

	err := r.getDb(ctx).
		Where("feature_id = ? AND valid_from IS NOT NULL AND EXISTS (?)", featureID, subQuery).
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}
