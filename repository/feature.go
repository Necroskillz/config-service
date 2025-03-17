package repository

import (
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type FeatureRepository struct {
	GormRepository[model.Feature]
}

func NewFeatureRepository(db *gorm.DB) *FeatureRepository {
	return &FeatureRepository{GormRepository[model.Feature]{db: db}}
}
