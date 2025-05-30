package variation

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/necroskillz/config-service/db"
)

type HierarchyService struct {
	queries *db.Queries
	cache   *ristretto.Cache[string, any]
}

func NewHierarchyService(queries *db.Queries, cache *ristretto.Cache[string, any]) *HierarchyService {
	return &HierarchyService{queries: queries, cache: cache}
}

type GetHierarchyConfig struct {
	ForceRefresh bool
}

type GetHierarchyConfigFunc func(config *GetHierarchyConfig)

func WithForceRefresh() GetHierarchyConfigFunc {
	return func(config *GetHierarchyConfig) {
		config.ForceRefresh = true
	}
}

const variationHierarchyCacheKey = "variation_hierarchy"

func (s *HierarchyService) GetVariationHierarchy(ctx context.Context, options ...GetHierarchyConfigFunc) (*Hierarchy, error) {
	config := GetHierarchyConfig{
		ForceRefresh: false,
	}

	for _, fn := range options {
		fn(&config)
	}

	if !config.ForceRefresh {
		cachedVariationHierarchy, exists := s.cache.Get(variationHierarchyCacheKey)

		if exists {
			return cachedVariationHierarchy.(*Hierarchy), nil
		}
	}

	variationPropertyValues, err := s.queries.GetVariationPropertyValues(ctx)
	if err != nil {
		return nil, err
	}

	serviceTypesProperties, err := s.queries.GetServiceTypeVariationProperties(ctx)
	if err != nil {
		return nil, err
	}

	variationHierarchy := NewHierarchy(variationPropertyValues, serviceTypesProperties)

	s.cache.SetWithTTL(variationHierarchyCacheKey, variationHierarchy, int64(len(variationPropertyValues)*10+len(serviceTypesProperties)), time.Minute*10)

	return variationHierarchy, nil
}

func (s *HierarchyService) ClearCache(ctx context.Context) {
	s.cache.Del(variationHierarchyCacheKey)
}
