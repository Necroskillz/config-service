package repository

import (
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ServiceTypeRepository struct {
	GormRepository[model.ServiceType]
}

func NewServiceTypeRepository(db *gorm.DB) *ServiceTypeRepository {
	return &ServiceTypeRepository{GormRepository: GormRepository[model.ServiceType]{db: db}}
}
