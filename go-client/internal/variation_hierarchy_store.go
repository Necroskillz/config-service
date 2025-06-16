package internal

import (
	"context"
	"fmt"
	"sync/atomic"
)

type VariationHierarchyStore struct {
	dataLoader         ConfigurationDataLoader
	config             *Config
	variationHierarchy atomic.Pointer[VariationHierarchy]
}

func (v *VariationHierarchyStore) storeVariationHierarchyFallbackFile(variationHierarchy *VariationHierarchy) error {
	if err := WriteFallbackFile(v.config, "variation_hierarchy.json", variationHierarchy); err != nil {
		return fmt.Errorf("failed to write variation hierarchy fallback file: %w", err)
	}

	return nil
}

func (v *VariationHierarchyStore) loadVariationHierarchyFallbackFile() (*VariationHierarchy, error) {
	variationHierarchy, err := ReadFallbackFile[VariationHierarchy](v.config, "variation_hierarchy.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read variation hierarchy fallback file: %w", err)
	}

	return variationHierarchy, nil
}

func NewVariationHierarchyStore(dataLoader ConfigurationDataLoader, config *Config) *VariationHierarchyStore {
	return &VariationHierarchyStore{
		dataLoader: dataLoader,
		config:     config,
	}
}

func (v *VariationHierarchyStore) Init(ctx context.Context) error {
	err := v.Refresh(ctx)

	if err != nil {
		if v.config.IsFallbackFileEnabled() {
			v.config.Logger.Error(ctx, "failed to get variation hierarchy, trying to load fallback file", "error", err)

			variationHierarchy, err := v.loadVariationHierarchyFallbackFile()
			if err != nil {
				return fmt.Errorf("failed to load variation hierarchy fallback file: %w", err)
			}

			v.variationHierarchy.Store(variationHierarchy)
		} else {
			return fmt.Errorf("failed to init variation hierarchy: %w", err)
		}
	}

	return nil
}

func (v *VariationHierarchyStore) GetVariationHierarchy(ctx context.Context) (*VariationHierarchy, error) {
	variationHierarchy := v.variationHierarchy.Load()
	if variationHierarchy == nil {
		return nil, fmt.Errorf("variation hierarchy store is not initialized")
	}

	return variationHierarchy, nil
}

func (v *VariationHierarchyStore) Refresh(ctx context.Context) error {
	h, err := v.dataLoader.GetVariationHierarchy(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh variation hierarchy: %w", err)
	}

	v.variationHierarchy.Store(h)

	if v.config.IsFallbackFileEnabled() {
		go func() {
			err := v.storeVariationHierarchyFallbackFile(h)
			if err != nil {
				v.config.Logger.Error(context.Background(), "failed to store variation hierarchy fallback file", "error", err)
			}
		}()
	}

	return nil
}
