package repository

import (
	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ChangesetChangeRepository struct {
	GormRepository[model.ChangesetChange]
}

func NewChangesetChangeRepository(db *gorm.DB) *ChangesetChangeRepository {
	return &ChangesetChangeRepository{
		GormRepository: GormRepository[model.ChangesetChange]{db},
	}
}
