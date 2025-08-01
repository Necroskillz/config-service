package core

import (
	"context"
	"fmt"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
)

type Service struct {
	queries             *db.Queries
	currentUserAccessor *auth.CurrentUserAccessor
}

func NewService(queries *db.Queries, currentUserAccessor *auth.CurrentUserAccessor) *Service {
	return &Service{queries: queries, currentUserAccessor: currentUserAccessor}
}

func (s *Service) GetService(ctx context.Context, serviceID uint) (db.Service, error) {
	service, err := s.queries.GetService(ctx, serviceID)
	if err != nil {
		return service, NewDbError(err, "Service")
	}

	return service, nil
}

func (s *Service) GetFeature(ctx context.Context, featureID uint) (db.Feature, error) {
	feature, err := s.queries.GetFeature(ctx, featureID)
	if err != nil {
		return feature, NewDbError(err, "Feature")
	}

	return feature, nil
}

func (s *Service) GetServiceVersion(ctx context.Context, serviceVersionID uint) (db.GetServiceVersionRow, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersion, err := s.queries.GetServiceVersion(ctx, db.GetServiceVersionParams{
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return serviceVersion, NewDbError(err, "ServiceVersion")
	}

	return serviceVersion, nil
}

func (s *Service) GetFeatureVersionWithoutLink(ctx context.Context, featureVersionID uint) (db.GetFeatureVersionRow, error) {
	user := s.currentUserAccessor.GetUser(ctx)
	featureVersion, err := s.queries.GetFeatureVersion(ctx, db.GetFeatureVersionParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return featureVersion, NewDbError(err, "FeatureVersion")
	}

	return featureVersion, nil
}

func (s *Service) GetFeatureVersionWithLink(ctx context.Context, serviceVersionID uint, featureVersionID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, db.GetFeatureVersionServiceVersionLinkRow, error) {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var link db.GetFeatureVersionServiceVersionLinkRow
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersion, err := s.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return serviceVersion, featureVersion, link, err
	}

	featureVersion, err = s.queries.GetFeatureVersion(ctx, db.GetFeatureVersionParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return serviceVersion, featureVersion, link, NewDbError(err, "FeatureVersion")
	}

	link, err = s.queries.GetFeatureVersionServiceVersionLink(ctx, db.GetFeatureVersionServiceVersionLinkParams{
		FeatureVersionID: featureVersion.ID,
		ServiceVersionID: serviceVersion.ID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return serviceVersion, featureVersion, link, NewDbError(err, "FeatureVersionServiceVersion")
	}

	return serviceVersion, featureVersion, link, nil
}

func (s *Service) GetFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, error) {
	serviceVersion, featureVersion, _, err := s.GetFeatureVersionWithLink(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return serviceVersion, featureVersion, err
	}

	return serviceVersion, featureVersion, nil
}

func (s *Service) GetKey(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, db.GetKeyRow, error) {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.GetKeyRow
	user := s.currentUserAccessor.GetUser(ctx)
	serviceVersion, featureVersion, err := s.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return serviceVersion, featureVersion, key, err
	}

	key, err = s.queries.GetKey(ctx, db.GetKeyParams{
		KeyID:       keyID,
		ChangesetID: user.ChangesetID,
	})
	if err != nil {
		return serviceVersion, featureVersion, key, NewDbError(err, "Key")
	}

	if key.FeatureVersionID != featureVersion.ID {
		return serviceVersion, featureVersion, key, NewServiceError(ErrorCodeRecordNotFound, fmt.Sprintf("Key %d not found in feature version %d", key.ID, featureVersion.ID))
	}

	return serviceVersion, featureVersion, key, nil
}

func (s *Service) GetVariationValue(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint, variationValueID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, db.GetKeyRow, db.VariationValue, error) {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.GetKeyRow
	var variationValue db.VariationValue
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersion, featureVersion, key, err := s.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return serviceVersion, featureVersion, key, variationValue, err
	}

	variationValue, err = s.queries.GetVariationValue(ctx, db.GetVariationValueParams{
		VariationValueID: variationValueID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return serviceVersion, featureVersion, key, variationValue, NewDbError(err, "VariationValue")
	}

	return serviceVersion, featureVersion, key, variationValue, nil
}

func (s *Service) GetKeyOptional(ctx context.Context, serviceVersionID uint, featureVersionID *uint, keyID *uint) (db.GetServiceVersionRow, *db.GetFeatureVersionRow, *db.GetKeyRow, error) {
	if featureVersionID != nil && keyID != nil {
		serviceVersion, featureVersion, key, err := s.GetKey(ctx, serviceVersionID, *featureVersionID, *keyID)
		if err != nil {
			return serviceVersion, nil, nil, err
		}

		return serviceVersion, &featureVersion, &key, nil
	} else if featureVersionID != nil {
		serviceVersion, featureVersion, err := s.GetFeatureVersion(ctx, serviceVersionID, *featureVersionID)
		if err != nil {
			return serviceVersion, nil, nil, err
		}

		return serviceVersion, &featureVersion, nil, nil
	} else {
		serviceVersion, err := s.GetServiceVersion(ctx, serviceVersionID)
		if err != nil {
			return serviceVersion, nil, nil, err
		}

		return serviceVersion, nil, nil, nil
	}
}
