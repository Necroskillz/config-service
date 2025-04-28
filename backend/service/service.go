package service

import (
	"context"
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
}

func NewServiceService(
	queries *db.Queries,
	unitOfWorkRunner db.UnitOfWorkRunner,
	changesetService *ChangesetService,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *Validator,
	coreService *CoreService,
) *ServiceService {
	return &ServiceService{
		unitOfWorkRunner:    unitOfWorkRunner,
		queries:             queries,
		changesetService:    changesetService,
		currentUserAccessor: currentUserAccessor,
		validator:           validator,
		coreService:         coreService,
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
}

func (s *ServiceService) GetServiceVersion(ctx context.Context, id uint) (ServiceVersionDto, error) {
	serviceVersion, err := s.queries.GetServiceVersion(ctx, id)
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
	}, nil
}

func (s *ServiceService) GetServiceVersionsForServiceVersion(ctx context.Context, serviceVersionID uint) ([]VersionLinkDto, error) {
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

	result := make([]VersionLinkDto, len(serviceVersions))
	for i, serviceVersion := range serviceVersions {
		result[i] = VersionLinkDto{
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

func (s *ServiceService) GetServiceTypes(ctx context.Context) ([]db.ServiceType, error) {
	return s.queries.GetServiceTypes(ctx)
}

func (s *ServiceService) GetServiceType(ctx context.Context, id uint) (db.ServiceType, error) {
	serviceType, err := s.queries.GetServiceType(ctx, id)
	if err != nil {
		return db.ServiceType{}, NewDbError(err, "ServiceType")
	}

	return serviceType, nil
}

type CreateServiceParams struct {
	Name          string
	Description   string
	ServiceTypeID uint
}

func (s *ServiceService) validateCreateService(ctx context.Context, data CreateServiceParams) error {
	return s.validator.
		Validate(data.Name, "Name").MaxLength(100).Required().ServiceNameNotTaken().
		Validate(data.Description, "Description").MaxLength(500).Required().
		Validate(data.ServiceTypeID, "Service Type ID").Min(1).
		Error(ctx)
}

func (s *ServiceService) CreateService(ctx context.Context, data CreateServiceParams) (uint, error) {
	if !s.currentUserAccessor.GetUser(ctx).IsGlobalAdmin {
		return 0, NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create services")
	}

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

func (s *ServiceService) UpdateService(ctx context.Context, data UpdateServiceParams) error {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, data.ServiceVersionID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if user.GetPermissionForService(serviceVersion.ServiceID) < constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to update this service")
	}

	return s.queries.UpdateService(ctx, db.UpdateServiceParams{
		ServiceID:   serviceVersion.ServiceID,
		Description: data.Description,
	})
}

func (s *ServiceService) PublishServiceVersion(ctx context.Context, serviceVersionID uint) error {
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

	return s.queries.PublishServiceVersion(ctx, serviceVersionID)
}
