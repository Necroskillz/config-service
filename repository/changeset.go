package repository

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"gorm.io/gorm"
)

type ChangesetRepository struct {
	GormRepository[model.Changeset]
}

func NewChangesetRepository(db *gorm.DB) *ChangesetRepository {
	return &ChangesetRepository{
		GormRepository: GormRepository[model.Changeset]{db},
	}
}

func (r *ChangesetRepository) GetOpenChangesetForUser(ctx context.Context, userID uint) (*model.Changeset, error) {
	changeset := model.Changeset{
		UserID: userID,
		State:  model.ChangesetStateOpen,
	}

	err := r.getDb(ctx).Where("user_id = ? AND state = ?", userID, model.ChangesetStateOpen).First(&changeset).Error

	return NilIfNotFound(&changeset, err)
}
