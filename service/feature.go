package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type FeatureService struct {
	unitOfWorkCreator                      repository.UnitOfWorkCreator
	featureRepository                      *repository.FeatureRepository
	featureVersionRepository               *repository.FeatureVersionRepository
	serviceVersionFeatureVersionRepository *repository.ServiceVersionFeatureVersionRepository
	changesetService                       *ChangesetService
}

func NewFeatureService(
	unitOfWorkCreator repository.UnitOfWorkCreator,
	featureRepository *repository.FeatureRepository,
	featureVersionRepository *repository.FeatureVersionRepository,
	serviceVersionFeatureVersionRepository *repository.ServiceVersionFeatureVersionRepository,
	changesetService *ChangesetService,
) *FeatureService {
	return &FeatureService{
		unitOfWorkCreator:                      unitOfWorkCreator,
		featureRepository:                      featureRepository,
		featureVersionRepository:               featureVersionRepository,
		serviceVersionFeatureVersionRepository: serviceVersionFeatureVersionRepository,
		changesetService:                       changesetService,
	}
}

func (s *FeatureService) GetFeatureByName(ctx context.Context, name string) (*model.Feature, error) {
	return s.featureRepository.GetByProperty(ctx, "name", name)
}

func (s *FeatureService) GetFeatureVersion(ctx context.Context, featureVersionID uint) (*model.FeatureVersion, error) {
	return s.featureVersionRepository.GetById(ctx, featureVersionID, "Feature")
}

func (s *FeatureService) GetFeatureVersionsLinkedToServiceVersion(ctx context.Context, featureID uint, serviceVersionID uint) ([]model.FeatureVersion, error) {
	featureVersions, err := s.featureVersionRepository.GetByFeatureIDForServiceVersion(ctx, featureID, serviceVersionID)

	if err != nil {
		return nil, err
	}

	return featureVersions, nil
}

func (s *FeatureService) GetServiceFeatures(ctx context.Context, serviceVersionID uint) ([]model.FeatureVersion, error) {
	links, err := s.serviceVersionFeatureVersionRepository.GetActive(ctx, serviceVersionID)

	if err != nil {
		return nil, err
	}

	featureVersions := make([]model.FeatureVersion, len(links))

	for i, link := range links {
		featureVersions[i] = link.FeatureVersion
	}

	return featureVersions, nil
}

type CreateFeatureParams struct {
	ChangesetID      uint
	ServiceVersionID uint
	Name             string
	Description      string
	ServiceID        uint
}

func (s *FeatureService) CreateFeature(ctx context.Context, params CreateFeatureParams) (*model.Feature, error) {
	feature := model.Feature{
		Name:        params.Name,
		Description: params.Description,
		ServiceID:   params.ServiceID,
	}

	return &feature, s.unitOfWorkCreator.Run(ctx, func(ctx context.Context) error {
		if err := s.featureRepository.Create(ctx, &feature); err != nil {
			return err
		}

		featureVersion := model.FeatureVersion{
			FeatureID: feature.ID,
			Version:   1,
		}

		if err := s.featureVersionRepository.Create(ctx, &featureVersion); err != nil {
			return err
		}

		if err := s.changesetService.AddFeatureVersionChange(ctx, params.ChangesetID, featureVersion.ID, model.ChangesetChangeTypeCreate); err != nil {
			return err
		}

		link := model.FeatureVersionServiceVersion{
			ServiceVersionID: params.ServiceVersionID,
			FeatureVersionID: featureVersion.ID,
		}

		if err := s.serviceVersionFeatureVersionRepository.Create(ctx, &link); err != nil {
			return err
		}

		if err := s.changesetService.AddFeatureVersionServiceVersionLinkChange(ctx, params.ChangesetID, link.ID, params.ServiceVersionID, featureVersion.ID, model.ChangesetChangeTypeCreate); err != nil {
			return err
		}

		return nil
	})
}
