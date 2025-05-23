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
)

type ContextService struct {
	queries          *db.Queries
	unitOfWorkRunner db.UnitOfWorkRunner
	cache            *ristretto.Cache[string, any]
}

func NewContextService(queries *db.Queries, unitOfWorkRunner db.UnitOfWorkRunner, cache *ristretto.Cache[string, any]) *ContextService {
	return &ContextService{
		queries:          queries,
		unitOfWorkRunner: unitOfWorkRunner,
		cache:            cache,
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

func (v *ContextService) GetVariationContextValues(ctx context.Context, variationContextID uint) (map[uint]string, error) {
	valuesCacheKey := getVariationContextValuesCacheKey(variationContextID)
	cachedValues, exists := v.cache.Get(valuesCacheKey)

	if exists {
		return cachedValues.(map[uint]string), nil
	}

	variationContextValues, err := v.queries.GetVariationContextValues(ctx, variationContextID)
	if err != nil {
		return nil, err
	}

	variationContext := make(map[uint]string, len(variationContextValues))
	valueIds := make([]uint, len(variationContextValues))
	for i, variationContextValue := range variationContextValues {
		variationContext[variationContextValue.PropertyID] = variationContextValue.Value
		valueIds[i] = variationContextValue.ValueID
	}

	v.cache.Set(valuesCacheKey, variationContext, int64(len(valueIds)*3))
	v.cache.Set(getVariationContextIdCacheKey(valueIds), variationContextID, 1)

	return variationContext, nil
}

func (v *ContextService) GetVariationContextID(ctx context.Context, variationContextValues []uint) (uint, error) {
	cacheKey := getVariationContextIdCacheKey(variationContextValues)
	cachedID, exists := v.cache.Get(cacheKey)

	if exists {
		return cachedID.(uint), nil
	}

	var contextID uint
	err := v.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		id, err := tx.GetVariationContextId(ctx, db.GetVariationContextIdParams{
			VariationPropertyValueIds: variationContextValues,
			PropertyCount:             len(variationContextValues),
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				newContextID, err := tx.CreateVariationContext(ctx)
				if err != nil {
					return err
				}

				for _, valueID := range variationContextValues {
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

	v.cache.Set(getVariationContextIdCacheKey(variationContextValues), contextID, 1)

	return contextID, nil
}
