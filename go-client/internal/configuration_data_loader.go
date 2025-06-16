package internal

import (
	"context"
	"fmt"

	grpcgen "github.com/necroskillz/config-service/go-client/grpc/gen"
)

type ConfigurationDataLoader struct {
	configClient grpcgen.ConfigServiceClient
	config       *Config
}

func NewConfigurationDataLoader(configClient grpcgen.ConfigServiceClient, config *Config) *ConfigurationDataLoader {
	return &ConfigurationDataLoader{
		configClient: configClient,
		config:       config,
	}
}

func (c *ConfigurationDataLoader) GetConfiguration(ctx context.Context, changesetID *uint32) (*ConfigurationSnapshot, error) {
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

func (c *ConfigurationDataLoader) GetVariationHierarchy(ctx context.Context) (*VariationHierarchy, error) {
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

func (c *ConfigurationDataLoader) GetNextChangesets(ctx context.Context, services []string, afterChangesetID uint32) ([]uint32, error) {
	req := &grpcgen.GetNextChangesetsRequest{
		Services:         services,
		AfterChangesetId: afterChangesetID,
	}

	res, err := c.configClient.GetNextChangesets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get next changesets: %w", err)
	}

	return res.ChangesetIds, nil
}
