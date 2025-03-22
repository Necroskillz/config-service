package repository

import (
	"context"

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

func (r *ChangesetChangeRepository) DeleteByVariationValueID(ctx context.Context, valueID uint) error {
	return r.db.Where("new_variation_value_id = ? OR old_variation_value_id = ?", valueID, valueID).Delete(&model.ChangesetChange{}).Error
}

func (r *ChangesetChangeRepository) DeleteByKeyID(ctx context.Context, keyID uint) error {
	return r.db.Where("key_id = ?", keyID).Delete(&model.ChangesetChange{}).Error
}

func (r *ChangesetChangeRepository) DeleteByFeatureVersionID(ctx context.Context, featureVersionID uint) error {
	return r.db.Where("feature_version_id = ?", featureVersionID).Delete(&model.ChangesetChange{}).Error
}

func (r *ChangesetChangeRepository) DeleteByServiceVersionID(ctx context.Context, serviceVersionID uint) error {
	return r.db.Where("service_version_id = ?", serviceVersionID).Delete(&model.ChangesetChange{}).Error
}

func (r *ChangesetChangeRepository) DeleteByFeatureVersionServiceVersionLinkID(ctx context.Context, featureVersionServiceVersionID uint) error {
	return r.db.Where("feature_version_service_version_id = ?", featureVersionServiceVersionID).Delete(&model.ChangesetChange{}).Error
}
