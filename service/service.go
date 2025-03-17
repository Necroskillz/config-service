package service

import (
	"context"
	"fmt"
	"time"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type ServiceService struct {
	unitOfWorkCreator        repository.UnitOfWorkCreator
	serviceRepository        *repository.ServiceRepository
	serviceVersionRepository *repository.ServiceVersionRepository
	serviceTypeRepository    *repository.ServiceTypeRepository
}

func NewServiceService(unitOfWorkCreator repository.UnitOfWorkCreator, serviceRepository *repository.ServiceRepository, serviceVersionRepository *repository.ServiceVersionRepository, serviceTypeRepository *repository.ServiceTypeRepository) *ServiceService {
	return &ServiceService{
		unitOfWorkCreator:        unitOfWorkCreator,
		serviceRepository:        serviceRepository,
		serviceVersionRepository: serviceVersionRepository,
		serviceTypeRepository:    serviceTypeRepository,
	}
}

func (s *ServiceService) GetServiceVersion(ctx context.Context, id uint) (*model.ServiceVersion, error) {
	return s.serviceVersionRepository.GetById(ctx, id, "Service")
}

func (s *ServiceService) GetServiceVersions(ctx context.Context, serviceID uint) ([]model.ServiceVersion, error) {
	return s.serviceVersionRepository.GetByServiceID(ctx, serviceID)
}

func (s *ServiceService) GetCurrentServiceVersions(ctx context.Context) ([]model.ServiceVersion, error) {
	return s.serviceVersionRepository.GetActive(ctx)
}

func (s *ServiceService) GetServiceTypes(ctx context.Context) ([]model.ServiceType, error) {
	return s.serviceTypeRepository.GetAll(ctx)
}

func (s *ServiceService) GetServiceByName(ctx context.Context, name string) (*model.Service, error) {
	return s.serviceRepository.GetByProperty(ctx, "name", name)
}

func (s *ServiceService) GetPermissionForService(ctx context.Context, user *auth.User, seviceVersionId uint) (constants.PermissionLevel, error) {
	serviceVersion, err := s.GetServiceVersion(ctx, seviceVersionId)
	if err != nil {
		return constants.PermissionViewer, err
	}

	if serviceVersion == nil {
		return constants.PermissionViewer, fmt.Errorf("service version %d not found", seviceVersionId)
	}

	return user.GetPermissionForService(serviceVersion.Service.ID), nil
}

func (s *ServiceService) CreateService(ctx context.Context, name string, description string, serviceTypeID uint) error {
	return s.unitOfWorkCreator.Run(ctx, func(ctx context.Context) error {
		service := model.Service{
			Name:          name,
			Description:   description,
			ServiceTypeID: serviceTypeID,
		}

		err := s.serviceRepository.Create(ctx, &service)
		if err != nil {
			return err
		}

		validFrom := time.Now()

		serviceVersion := model.ServiceVersion{
			ServiceID: service.ID,
			Version:   1,
			ValidFrom: &validFrom,
		}

		err = s.serviceVersionRepository.Create(ctx, &serviceVersion)
		if err != nil {
			return err
		}

		return nil
	})
}
