package service

import (
	"context"

	"github.com/necroskillz/config-service/db"
)

type ServiceService struct {
	unitOfWorkRunner db.UnitOfWorkRunner
	queries          *db.Queries
}

func NewServiceService(queries *db.Queries, unitOfWorkRunner db.UnitOfWorkRunner) *ServiceService {
	return &ServiceService{
		unitOfWorkRunner: unitOfWorkRunner,
		queries:          queries,
	}
}

func (s *ServiceService) GetServiceVersion(ctx context.Context, id uint) (db.GetServiceVersionRow, error) {
	return s.queries.GetServiceVersion(ctx, id)
}

func (s *ServiceService) GetServiceVersions(ctx context.Context, serviceID uint, changesetID uint) ([]db.GetServiceVersionsForServiceRow, error) {
	return s.queries.GetServiceVersionsForService(ctx, db.GetServiceVersionsForServiceParams{
		ServiceID:   serviceID,
		ChangesetID: changesetID,
	})
}

func (s *ServiceService) GetCurrentServiceVersions(ctx context.Context, changesetID uint) ([]db.GetActiveServiceVersionsRow, error) {
	return s.queries.GetActiveServiceVersions(ctx, changesetID)
}

func (s *ServiceService) GetServiceTypes(ctx context.Context) ([]db.ServiceType, error) {
	return s.queries.GetServiceTypes(ctx)
}

func (s *ServiceService) CreateService(ctx context.Context, name string, description string, serviceTypeID uint, changesetID uint) (uint, error) {
	var serviceVersionId uint

	err := s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		serviceId, err := tx.CreateService(ctx, db.CreateServiceParams{
			Name:          name,
			Description:   description,
			ServiceTypeID: serviceTypeID,
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
