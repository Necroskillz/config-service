package repository

import (
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ServiceRepository struct {
	GormRepository[model.Service]
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{GormRepository[model.Service]{db: db}}
}
