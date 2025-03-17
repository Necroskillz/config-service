package repository

import (
	"context"
	"time"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type KeyRepository struct {
	GormRepository[model.Key]
}

func NewKeyRepository(db *gorm.DB) *KeyRepository {
	return &KeyRepository{GormRepository[model.Key]{db: db}}
}

func (r *KeyRepository) GetActive(ctx context.Context, featureVersionID uint) ([]model.Key, error) {
	var keys []model.Key

	err := r.getDb(ctx).
		Where("feature_version_id = ? AND valid_from <= ? AND (valid_to >= ? OR valid_to IS NULL)", featureVersionID, time.Now(), time.Now()).
		Find(&keys).Error

	return keys, err
}

func (r *KeyRepository) GetActiveKeyByName(ctx context.Context, featureVersionID uint, name string) (*model.Key, error) {
	var key model.Key

	err := r.getDb(ctx).
		Where("name = ? AND feature_version_id = ? AND valid_from <= ? AND (valid_to >= ? OR valid_to IS NULL)", name, featureVersionID, time.Now(), time.Now()).
		First(&key).Error

	return NilIfNotFound(&key, err)
}
