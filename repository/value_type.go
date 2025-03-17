package repository

import (
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ValueTypeRepository struct {
	GormRepository[model.ValueType]
}

func NewValueTypeRepository(db *gorm.DB) *ValueTypeRepository {
	return &ValueTypeRepository{GormRepository[model.ValueType]{db}}
}
