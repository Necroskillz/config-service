package service

import (
	"context"
	"fmt"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
)

type CoreService struct {
	queries             *db.Queries
	currentUserAccessor *auth.CurrentUserAccessor
}

func NewCoreService(queries *db.Queries, currentUserAccessor *auth.CurrentUserAccessor) *CoreService {
	return &CoreService{queries: queries, currentUserAccessor: currentUserAccessor}
}

func (s *CoreService) GetServiceVersion(ctx context.Context, serviceVersionID uint) (db.GetServiceVersionRow, error) {
	serviceVersion, err := s.queries.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return serviceVersion, NewDbError(err, "ServiceVersion")
	}

	return serviceVersion, nil
}

func (s *CoreService) GetFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, error) {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersion, err := s.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return serviceVersion, featureVersion, err
	}

	featureVersion, err = s.queries.GetFeatureVersion(ctx, featureVersionID)
	if err != nil {
		return serviceVersion, featureVersion, NewDbError(err, "FeatureVersion")
	}

	linked, err := s.queries.IsFeatureVersionLinkedToServiceVersion(ctx, db.IsFeatureVersionLinkedToServiceVersionParams{
		FeatureVersionID: featureVersion.ID,
		ServiceVersionID: serviceVersion.ID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return serviceVersion, featureVersion, NewDbError(err, "FeatureVersion")
	}

	if !linked {
		return serviceVersion, featureVersion, NewServiceError(ErrorCodeRecordNotFound, fmt.Sprintf("Feature version %d is not linked to service version %d", featureVersion.ID, serviceVersion.ID))
	}

	return serviceVersion, featureVersion, nil
}

func (s *CoreService) GetKey(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, db.GetKeyRow, error) {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.GetKeyRow

	serviceVersion, featureVersion, err := s.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return serviceVersion, featureVersion, key, err
	}

	key, err = s.queries.GetKey(ctx, keyID)
	if err != nil {
		return serviceVersion, featureVersion, key, NewDbError(err, "Key")
	}

	if key.FeatureVersionID != featureVersion.ID {
		return serviceVersion, featureVersion, key, NewServiceError(ErrorCodeRecordNotFound, fmt.Sprintf("Key %d not found in feature version %d", key.ID, featureVersion.ID))
	}

	return serviceVersion, featureVersion, key, nil
}

func (s *CoreService) GetVariationValue(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint, variationValueID uint) (db.GetServiceVersionRow, db.GetFeatureVersionRow, db.GetKeyRow, db.VariationValue, error) {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.GetKeyRow
	var variationValue db.VariationValue

	serviceVersion, featureVersion, key, err := s.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return serviceVersion, featureVersion, key, variationValue, err
	}

	variationValue, err = s.queries.GetVariationValue(ctx, variationValueID)
	if err != nil {
		return serviceVersion, featureVersion, key, variationValue, NewDbError(err, "VariationValue")
	}

	return serviceVersion, featureVersion, key, variationValue, nil
}
