package repository

import (
	"context"
	"time"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ServiceVersionFeatureVersionRepository struct {
	GormRepository[model.FeatureVersionServiceVersion]
}

func NewServiceVersionFeatureVersionRepository(db *gorm.DB) *ServiceVersionFeatureVersionRepository {
	return &ServiceVersionFeatureVersionRepository{
		GormRepository: GormRepository[model.FeatureVersionServiceVersion]{db},
	}
}

func (r *ServiceVersionFeatureVersionRepository) GetActive(ctx context.Context, serviceVersionID uint) ([]model.FeatureVersionServiceVersion, error) {
	var featureVersionServiceVersions []model.FeatureVersionServiceVersion

	err := r.getDb(ctx).Preload("FeatureVersion").Preload("FeatureVersion.Feature").
		Where("service_version_id = ? AND valid_from <= ? AND (valid_to >= ? OR valid_to IS NULL)", serviceVersionID, time.Now(), time.Now()).
		Find(&featureVersionServiceVersions).Error

	return featureVersionServiceVersions, err
}
