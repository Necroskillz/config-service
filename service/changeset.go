package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type ChangesetService struct {
	changesetRepository       *repository.ChangesetRepository
	changesetChangeRepository *repository.ChangesetChangeRepository
}

func NewChangesetService(changesetRepository *repository.ChangesetRepository, changesetChangeRepository *repository.ChangesetChangeRepository) *ChangesetService {
	return &ChangesetService{changesetRepository: changesetRepository, changesetChangeRepository: changesetChangeRepository}
}

func (s *ChangesetService) GetOpenChangesetForUser(ctx context.Context, userID uint) (*model.Changeset, error) {
	return s.changesetRepository.GetOpenChangesetForUser(ctx, userID)
}

func (s *ChangesetService) CreateChangesetForUser(ctx context.Context, userID uint) (*model.Changeset, error) {
	changeset := &model.Changeset{
		UserID: userID,
		State:  model.ChangesetStateOpen,
	}

	err := s.changesetRepository.Create(ctx, changeset)

	if err != nil {
		return nil, err
	}

	return changeset, nil
}

func (s *ChangesetService) GetChangeset(ctx context.Context, changesetID uint) (*model.Changeset, error) {
	return s.changesetRepository.GetById(
		ctx,
		changesetID,
		"ChangesetChanges",
		"ChangesetChanges.ServiceVersion",
		"ChangesetChanges.FeatureVersion",
		"ChangesetChanges.Key",
		"ChangesetChanges.ServiceVersion.Service",
		"ChangesetChanges.FeatureVersion.Feature",
		"ChangesetChanges.FeatureVersionServiceVersion",
		"ChangesetChanges.VariationValue",
		"ChangesetChanges.VariationValue.VariationPropertyValues",
	)
}

func (s *ChangesetService) AddCreateServiceVersionChange(ctx context.Context, changesetID uint, serviceVersionID uint) error {
	return s.changesetChangeRepository.Create(ctx, &model.ChangesetChange{
		ChangesetID:      changesetID,
		ServiceVersionID: &serviceVersionID,
		Type:             model.ChangesetChangeTypeCreate,
	})
}

func (s *ChangesetService) AddFeatureVersionChange(ctx context.Context, changesetID uint, featureVersionID uint, changeType model.ChangesetChangeType) error {
	return s.changesetChangeRepository.Create(ctx, &model.ChangesetChange{
		ChangesetID:      changesetID,
		FeatureVersionID: &featureVersionID,
		Type:             changeType,
	})
}

func (s *ChangesetService) AddFeatureVersionServiceVersionLinkChange(ctx context.Context, changesetID uint, featureVersionServiceVersionID uint, serviceVersionID uint, featureVersionID uint, changeType model.ChangesetChangeType) error {
	return s.changesetChangeRepository.Create(ctx, &model.ChangesetChange{
		ChangesetID:                    changesetID,
		FeatureVersionServiceVersionID: &featureVersionServiceVersionID,
		ServiceVersionID:               &serviceVersionID,
		FeatureVersionID:               &featureVersionID,
		Type:                           changeType,
	})
}

func (s *ChangesetService) AddKeyChange(ctx context.Context, changesetID uint, featureVersionID uint, keyID uint, changeType model.ChangesetChangeType) error {
	return s.changesetChangeRepository.Create(ctx, &model.ChangesetChange{
		ChangesetID:      changesetID,
		FeatureVersionID: &featureVersionID,
		KeyID:            &keyID,
		Type:             changeType,
	})
}

func (s *ChangesetService) AddVariationValueChange(ctx context.Context, changesetID uint, featureVersionID uint, keyID uint, variationValueID uint, changeType model.ChangesetChangeType, newValue *string, oldValue *string) error {
	return s.changesetChangeRepository.Create(ctx, &model.ChangesetChange{
		ChangesetID:      changesetID,
		FeatureVersionID: &featureVersionID,
		KeyID:            &keyID,
		VariationValueID: &variationValueID,
		Type:             changeType,
		NewValue:         newValue,
		OldValue:         oldValue,
	})
}
