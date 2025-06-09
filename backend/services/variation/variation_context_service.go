package variation

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
)

type ContextService struct {
	queries                   *db.Queries
	variationHierarchyService *HierarchyService
	unitOfWorkRunner          db.UnitOfWorkRunner
	cache                     *ristretto.Cache[string, any]
}

func NewContextService(queries *db.Queries, variationHierarchyService *HierarchyService, unitOfWorkRunner db.UnitOfWorkRunner, cache *ristretto.Cache[string, any]) *ContextService {
	return &ContextService{
		queries:                   queries,
		variationHierarchyService: variationHierarchyService,
		unitOfWorkRunner:          unitOfWorkRunner,
		cache:                     cache,
	}
}

func getVariationContextIdCacheKey(variationValues []uint) string {
	slices.Sort(variationValues)

	var sb strings.Builder
	sb.WriteString("variation_context_id")
	for _, valueID := range variationValues {
		sb.WriteString(fmt.Sprintf(":%d", valueID))
	}

	return sb.String()
}

func getVariationContextValuesCacheKey(variationContextID uint) string {
	return fmt.Sprintf("variation_context_values:%d", variationContextID)
}

func (s *ContextService) getIDsFromVariation(ctx context.Context, variation map[uint]string) ([]uint, error) {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]uint, 0, len(variation))
	for propertyID, value := range variation {
		value, err := variationHierarchy.GetPropertyValue(propertyID, value)
		if err != nil {
			return nil, err
		}

		if value.Archived {
			return nil, core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Value %s for property with ID %d is archived", value.Value, propertyID))
		}

		ids = append(ids, value.ID)
	}

	return ids, nil
}

func (s *ContextService) GetVariationContextValues(ctx context.Context, variationContextID uint) (map[uint]string, error) {
	valuesCacheKey := getVariationContextValuesCacheKey(variationContextID)
	cachedValues, exists := s.cache.Get(valuesCacheKey)

	if exists {
		return cachedValues.(map[uint]string), nil
	}

	variationContextValues, err := s.queries.GetVariationContextValues(ctx, variationContextID)
	if err != nil {
		return nil, err
	}

	variationContext := make(map[uint]string, len(variationContextValues))
	valueIds := make([]uint, len(variationContextValues))
	for i, variationContextValue := range variationContextValues {
		variationContext[variationContextValue.PropertyID] = variationContextValue.Value
		valueIds[i] = variationContextValue.ValueID
	}

	s.cache.Set(valuesCacheKey, variationContext, int64(len(valueIds)*3))
	s.cache.Set(getVariationContextIdCacheKey(valueIds), variationContextID, 1)

	return variationContext, nil
}

func (s *ContextService) GetVariationContextID(ctx context.Context, variation map[uint]string) (uint, error) {
	ids, err := s.getIDsFromVariation(ctx, variation)
	if err != nil {
		return 0, err
	}

	cacheKey := getVariationContextIdCacheKey(ids)
	cachedID, exists := s.cache.Get(cacheKey)

	if exists {
		return cachedID.(uint), nil
	}

	var contextID uint
	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		id, err := tx.GetVariationContextID(ctx, db.GetVariationContextIDParams{
			VariationPropertyValueIds: ids,
			PropertyCount:             len(ids),
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				newContextID, err := tx.CreateVariationContext(ctx)
				if err != nil {
					return err
				}

				for _, valueID := range ids {
					if err := tx.CreateVariationContextValue(ctx, db.CreateVariationContextValueParams{
						VariationContextID:       newContextID,
						VariationPropertyValueID: valueID,
					}); err != nil {
						return err
					}
				}

				contextID = newContextID
				return nil
			}

			return err
		}

		contextID = id
		return nil
	})

	if err != nil {
		return 0, err
	}

	s.cache.Set(cacheKey, contextID, 1)

	return contextID, nil
}
