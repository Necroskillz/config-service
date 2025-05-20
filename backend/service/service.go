package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type ServiceService struct {
	unitOfWorkRunner    db.UnitOfWorkRunner
	queries             *db.Queries
	changesetService    *ChangesetService
	currentUserAccessor *auth.CurrentUserAccessor
	validator           *Validator
	coreService         *CoreService
	validationService   *ValidationService
}

func NewServiceService(
	queries *db.Queries,
	unitOfWorkRunner db.UnitOfWorkRunner,
	changesetService *ChangesetService,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *Validator,
	coreService *CoreService,
	validationService *ValidationService,
) *ServiceService {
	return &ServiceService{
		unitOfWorkRunner:    unitOfWorkRunner,
		queries:             queries,
		changesetService:    changesetService,
		currentUserAccessor: currentUserAccessor,
		validator:           validator,
		coreService:         coreService,
		validationService:   validationService,
	}
}

type ServiceVersionDto struct {
	ID              uint   `json:"id" validate:"required"`
	ServiceID       uint   `json:"serviceId" validate:"required"`
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description" validate:"required"`
	Version         int    `json:"version" validate:"required"`
	Published       bool   `json:"published" validate:"required"`
	CanEdit         bool   `json:"canEdit" validate:"required"`
	ServiceTypeID   uint   `json:"serviceTypeId" validate:"required"`
	ServiceTypeName string `json:"serviceTypeName" validate:"required"`
	IsLastVersion   bool   `json:"isLastVersion" validate:"required"`
}

func (s *ServiceService) GetServiceVersion(ctx context.Context, id uint) (ServiceVersionDto, error) {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, id)
	if err != nil {
		return ServiceVersionDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	return ServiceVersionDto{
		ID:              serviceVersion.ID,
		ServiceID:       serviceVersion.ServiceID,
		Name:            serviceVersion.ServiceName,
		Description:     serviceVersion.ServiceDescription,
		Version:         serviceVersion.Version,
		Published:       serviceVersion.Published,
		CanEdit:         user.GetPermissionForService(serviceVersion.ServiceID) >= constants.PermissionAdmin,
		ServiceTypeID:   serviceVersion.ServiceTypeID,
		ServiceTypeName: serviceVersion.ServiceTypeName,
		IsLastVersion:   serviceVersion.LastVersion == serviceVersion.Version,
	}, nil
}

type ServiceVersionLinkDto struct {
	ID      uint `json:"id" validate:"required"`
	Version int  `json:"version" validate:"required"`
}

func (s *ServiceService) GetServiceVersionsForServiceVersion(ctx context.Context, serviceVersionID uint) ([]ServiceVersionLinkDto, error) {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersions, err := s.queries.GetServiceVersionsForService(ctx, db.GetServiceVersionsForServiceParams{
		ServiceID:   serviceVersion.ServiceID,
		ChangesetID: user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]ServiceVersionLinkDto, len(serviceVersions))
	for i, serviceVersion := range serviceVersions {
		result[i] = ServiceVersionLinkDto{
			ID:      serviceVersion.ID,
			Version: serviceVersion.Version,
		}
	}

	return result, nil
}

type ServiceVersionInfoDto struct {
	ID        uint `json:"id" validate:"required"`
	Published bool `json:"published" validate:"required"`
	Version   int  `json:"version" validate:"required"`
}

type ServiceDto struct {
	Name        string                  `json:"name" validate:"required"`
	Description string                  `json:"description" validate:"required"`
	Versions    []ServiceVersionInfoDto `json:"versions" validate:"required"`
}

func (s *ServiceService) GetServices(ctx context.Context) ([]ServiceDto, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersions, err := s.queries.GetServiceVersions(ctx, user.ChangesetID)
	if err != nil {
		return nil, err
	}

	services := make(map[uint]ServiceDto)

	for _, serviceVersion := range serviceVersions {
		if service, ok := services[serviceVersion.ServiceID]; ok {
			// display the last published and draft ones after the last published
			if serviceVersion.Published {
				service.Versions = []ServiceVersionInfoDto{
					{
						ID:        serviceVersion.ID,
						Published: true,
						Version:   serviceVersion.Version,
					},
				}
			} else {
				service.Versions = append(service.Versions, ServiceVersionInfoDto{
					ID:        serviceVersion.ID,
					Published: false,
					Version:   serviceVersion.Version,
				})
			}

			services[serviceVersion.ServiceID] = service
		} else {
			services[serviceVersion.ServiceID] = ServiceDto{
				Name:        serviceVersion.ServiceName,
				Description: serviceVersion.ServiceDescription,
				Versions: []ServiceVersionInfoDto{
					{
						ID:        serviceVersion.ID,
						Published: serviceVersion.Published,
						Version:   serviceVersion.Version,
					},
				},
			}
		}
	}

	result := make([]ServiceDto, 0, len(services))
	for _, service := range services {
		result = append(result, service)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)) < 0
	})

	return result, nil
}

type CreateServiceParams struct {
	Name          string
	Description   string
	ServiceTypeID uint
}

func (s *ServiceService) validateCreateService(ctx context.Context, data CreateServiceParams) error {
	if !s.currentUserAccessor.GetUser(ctx).IsGlobalAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create services")
	}

	err := s.validator.
		Validate(data.Name, "Name").Required().MaxLength(100).Regex(`^[\w\-_\.]+$`).
		Validate(data.Description, "Description").Func(optionalDescriptionValidatorFunc).
		Validate(data.ServiceTypeID, "Service Type ID").Min(1).
		Error(ctx)

	if err != nil {
		return err
	}

	if taken, err := s.validationService.IsServiceNameTaken(ctx, data.Name); err != nil {
		return err
	} else if taken {
		return NewServiceError(ErrorCodeInvalidOperation, "Service name is already taken")
	}

	return nil
}

func (s *ServiceService) CreateService(ctx context.Context, data CreateServiceParams) (uint, error) {
	if err := s.validateCreateService(ctx, data); err != nil {
		return 0, err
	}

	var serviceVersionId uint

	err := s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		serviceId, err := tx.CreateService(ctx, db.CreateServiceParams{
			Name:          data.Name,
			Description:   data.Description,
			ServiceTypeID: data.ServiceTypeID,
		})
		if err != nil {
			return err
		}

		serviceVersionId, err = tx.CreateServiceVersion(ctx, db.CreateServiceVersionParams{
			ServiceID: serviceId,
			Version:   1,
		})
		if err != nil {
			return err
		}

		tx.AddCreateServiceVersionChange(ctx, db.AddCreateServiceVersionChangeParams{
			ChangesetID:      changesetID,
			ServiceVersionID: serviceVersionId,
		})

		return nil
	})

	if err != nil {
		return 0, err
	}

	return serviceVersionId, nil
}

type UpdateServiceParams struct {
	ServiceVersionID uint
	Description      string
}

func (s *ServiceService) validateUpdateService(ctx context.Context, data UpdateServiceParams, serviceVersion db.GetServiceVersionRow) error {
	err := s.validator.
		Validate(data.Description, "Description").Func(descriptionValidatorFunc).
		Error(ctx)

	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if user.GetPermissionForService(serviceVersion.ServiceID) < constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to update this service")
	}

	return nil
}

func (s *ServiceService) UpdateService(ctx context.Context, data UpdateServiceParams) error {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, data.ServiceVersionID)
	if err != nil {
		return err
	}

	if err := s.validateUpdateService(ctx, data, serviceVersion); err != nil {
		return err
	}

	return s.queries.UpdateService(ctx, db.UpdateServiceParams{
		ServiceID:   serviceVersion.ServiceID,
		Description: data.Description,
	})
}

func (s *ServiceService) validateCreateServiceVersion(serviceVersion db.GetServiceVersionRow, user *auth.User) error {
	if serviceVersion.LastVersion != serviceVersion.Version {
		return NewServiceError(ErrorCodeInvalidOperation, "New service version can only be created from the latest version")
	}

	if user.GetPermissionForService(serviceVersion.ServiceID) < constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create a new service version")
	}

	return nil
}

func (s *ServiceService) CreateServiceVersion(ctx context.Context, serviceVersionID uint) (uint, error) {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return 0, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if err := s.validateCreateServiceVersion(serviceVersion, user); err != nil {
		return 0, err
	}

	var serviceVersionId uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		serviceVersionId, err = tx.CreateServiceVersion(ctx, db.CreateServiceVersionParams{
			ServiceID: serviceVersion.ServiceID,
			Version:   serviceVersion.Version + 1,
		})
		if err != nil {
			return err
		}

		tx.AddCreateServiceVersionChange(ctx, db.AddCreateServiceVersionChangeParams{
			ChangesetID:              changesetID,
			ServiceVersionID:         serviceVersionId,
			PreviousServiceVersionID: &serviceVersion.ID,
		})

		featureVersions, err := tx.GetFeatureVersionsForServiceVersion(ctx, db.GetFeatureVersionsForServiceVersionParams{
			ServiceVersionID: serviceVersionID,
			ChangesetID:      user.ChangesetID,
		})
		if err != nil {
			return err
		}

		for _, featureVersion := range featureVersions {
			linkId, err := tx.CreateFeatureVersionServiceVersion(ctx, db.CreateFeatureVersionServiceVersionParams{
				FeatureVersionID: featureVersion.ID,
				ServiceVersionID: serviceVersionId,
			})
			if err != nil {
				return err
			}

			tx.AddCreateFeatureVersionServiceVersionChange(ctx, db.AddCreateFeatureVersionServiceVersionChangeParams{
				ChangesetID:                    changesetID,
				FeatureVersionID:               featureVersion.ID,
				ServiceVersionID:               serviceVersionId,
				FeatureVersionServiceVersionID: linkId,
			})
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return serviceVersionId, nil
}

func (s *ServiceService) validatePublishServiceVersion(ctx context.Context, serviceVersionID uint) error {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if user.GetPermissionForService(serviceVersion.ServiceID) < constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to publish this service")
	}

	if serviceVersion.Published {
		return NewServiceError(ErrorCodeInvalidOperation, "This service version is already published")
	}

	changesCount, err := s.queries.GetRelatedServiceVersionChangesCount(ctx, db.GetRelatedServiceVersionChangesCountParams{
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return err
	}

	if changesCount > 0 {
		return NewServiceError(ErrorCodeInvalidOperation,
			fmt.Sprintf("Your current changeset contains %d changes related to this service version. Please apply or discard them before publishing.", changesCount),
		)
	}

	return nil
}

func (s *ServiceService) PublishServiceVersion(ctx context.Context, serviceVersionID uint) error {
	if err := s.validatePublishServiceVersion(ctx, serviceVersionID); err != nil {
		return err
	}

	return s.queries.PublishServiceVersion(ctx, serviceVersionID)
}
