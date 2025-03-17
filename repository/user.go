package repository

import (
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	GormRepository[model.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{GormRepository[model.User]{db: db}}
}
