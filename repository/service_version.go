package repository

import (
	"context"
	"time"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ServiceVersionRepository struct {
	GormRepository[model.ServiceVersion]
}

func NewServiceVersionRepository(db *gorm.DB) *ServiceVersionRepository {
	return &ServiceVersionRepository{GormRepository[model.ServiceVersion]{db: db}}
}

func (r *ServiceVersionRepository) GetActive(ctx context.Context) ([]model.ServiceVersion, error) {
	var result []model.ServiceVersion

	err := r.getDb(ctx).Preload("Service").
		Where("valid_from <= ? AND (valid_to >= ? OR valid_to IS NULL)", time.Now(), time.Now()).
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *ServiceVersionRepository) GetByServiceID(ctx context.Context, serviceID uint) ([]model.ServiceVersion, error) {
	var result []model.ServiceVersion

	err := r.getDb(ctx).
		Where("service_id = ? AND valid_from IS NOT NULL", serviceID).
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}
