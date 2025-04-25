package service

import (
	"context"

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
}

func NewServiceService(queries *db.Queries, unitOfWorkRunner db.UnitOfWorkRunner, changesetService *ChangesetService, currentUserAccessor *auth.CurrentUserAccessor, validator *Validator) *ServiceService {
	return &ServiceService{
		unitOfWorkRunner:    unitOfWorkRunner,
		queries:             queries,
		changesetService:    changesetService,
		currentUserAccessor: currentUserAccessor,
		validator:           validator,
	}
}

type ServiceVersionDto struct {
	ID            uint   `json:"id" validate:"required"`
	ServiceID     uint   `json:"serviceId" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description" validate:"required"`
	Version       int    `json:"version" validate:"required"`
	Published     bool   `json:"published" validate:"required"`
	CanEdit       bool   `json:"canEdit" validate:"required"`
	ServiceTypeID uint   `json:"serviceTypeId" validate:"required"`
}

func (s *ServiceService) GetServiceVersion(ctx context.Context, id uint) (ServiceVersionDto, error) {
	serviceVersion, err := s.queries.GetServiceVersion(ctx, id)
	if err != nil {
		return ServiceVersionDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	return ServiceVersionDto{
		ID:            serviceVersion.ID,
		ServiceID:     serviceVersion.ServiceID,
		Name:          serviceVersion.ServiceName,
		Description:   serviceVersion.ServiceDescription,
		Version:       serviceVersion.Version,
		Published:     serviceVersion.Published,
		CanEdit:       user.GetPermissionForService(serviceVersion.ServiceID) >= constants.PermissionAdmin,
		ServiceTypeID: serviceVersion.ServiceTypeID,
	}, nil
}

func (s *ServiceService) GetServiceVersions(ctx context.Context, serviceID uint) ([]VersionLinkDto, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersions, err := s.queries.GetServiceVersionsForService(ctx, db.GetServiceVersionsForServiceParams{
		ServiceID:   serviceID,
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

func (s *ServiceService) GetCurrentServiceVersions(ctx context.Context) ([]ServiceVersionDto, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersions, err := s.queries.GetActiveServiceVersions(ctx, user.ChangesetID)
	if err != nil {
		return nil, err
	}

	result := make([]ServiceVersionDto, len(serviceVersions))
	for i, serviceVersion := range serviceVersions {
		result[i] = ServiceVersionDto{
			ID:            serviceVersion.ID,
			ServiceID:     serviceVersion.ServiceID,
			Name:          serviceVersion.ServiceName,
			Description:   serviceVersion.ServiceDescription,
			Version:       serviceVersion.Version,
			Published:     serviceVersion.Published,
			CanEdit:       user.GetPermissionForService(serviceVersion.ServiceID) >= constants.PermissionAdmin,
			ServiceTypeID: serviceVersion.ServiceTypeID,
		}
	}

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
