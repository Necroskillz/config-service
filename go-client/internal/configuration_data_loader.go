package internal

import (
	"context"
	"fmt"

	grpcgen "github.com/necroskillz/config-service/go-client/grpc/gen"
)

var _ ConfigurationDataLoader = (*ConfigurationDataLoaderImpl)(nil)

type ConfigurationDataLoader interface {
	GetConfiguration(ctx context.Context, changesetID *uint32) (*ConfigurationSnapshot, error)
	GetVariationHierarchy(ctx context.Context) (*VariationHierarchy, error)
	GetNextChangesets(ctx context.Context, afterChangesetID uint32) ([]uint32, error)
}

type ConfigurationDataLoaderImpl struct {
	configClient grpcgen.ConfigServiceClient
	config       *Config
}

func NewConfigurationDataLoader(configClient grpcgen.ConfigServiceClient, config *Config) ConfigurationDataLoader {
	return &ConfigurationDataLoaderImpl{
		configClient: configClient,
		config:       config,
	}
}

func (c *ConfigurationDataLoaderImpl) GetConfiguration(ctx context.Context, changesetID *uint32) (*ConfigurationSnapshot, error) {
	mode := ""
	if c.config.ProductionMode {
		mode = "production"
	}

	req := &grpcgen.GetConfigurationRequest{
		Services:    c.config.Services,
		Variation:   c.config.StaticVariation,
		Mode:        &mode,
		ChangesetId: changesetID,
	}

	res, err := c.configClient.GetConfiguration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	snapshot := NewConfigurationSnapshot(res)

	snapshot.Validate(c.config.Features)

	return snapshot, nil
}

func (c *ConfigurationDataLoaderImpl) GetVariationHierarchy(ctx context.Context) (*VariationHierarchy, error) {
	req := &grpcgen.GetVariationHierarchyRequest{
		Services: c.config.Services,
	}

	res, err := c.configClient.GetVariationHierarchy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get variation hierarchy: %w", err)
	}

	variationHierarchy := NewVariationHierarchy(res)

	if err := variationHierarchy.Validate(c.config.StaticVariation, c.config.DynamicVariationResolvers); err != nil {
		return nil, fmt.Errorf("variation hierarchy is invalid: %w", err)
	}

	return variationHierarchy, nil
}

func (c *ConfigurationDataLoaderImpl) GetNextChangesets(ctx context.Context, afterChangesetID uint32) ([]uint32, error) {
	req := &grpcgen.GetNextChangesetsRequest{
		Services:         c.config.Services,
		AfterChangesetId: afterChangesetID,
	}

	res, err := c.configClient.GetNextChangesets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get next changesets: %w", err)
	}

	return res.ChangesetIds, nil
}
