package internal

import (
	"context"
	"time"
)

type Feature interface {
	FeatureName() string
}

type PropertyResolverFunc func(ctx context.Context) (string, error)

type Overrides map[string]map[string]any

func (o Overrides) Get(feature string, key string) (any, bool) {
	if _, ok := o[feature]; !ok {
		return nil, false
	}

	if _, ok := o[feature][key]; !ok {
		return nil, false
	}

	return o[feature][key], true
}

func (o Overrides) Set(feature string, key string, value any) {
	if _, ok := o[feature]; !ok {
		o[feature] = make(map[string]any)
	}

	o[feature][key] = value
}

type Config struct {
	Url                       string
	Services                  []string
	StaticVariation           map[string]string
	DynamicVariationResolvers map[string]PropertyResolverFunc
	Features                  []Feature
	ProductionMode            bool
	ChangesetOverrider        func(ctx context.Context) *uint32
	Logger                    Logger
	PollingInterval           time.Duration
	SnapshotCleanupInterval   time.Duration
	UnusedSnapshotExpiration  time.Duration
	FallbackFileLocation      string
	Overrides                 Overrides
}

func (c *Config) IsFallbackFileEnabled() bool {
	return c.FallbackFileLocation != ""
}
