package configserviceclient

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	grpcgen "github.com/necroskillz/config-service/go-client/grpc/gen"
	"github.com/necroskillz/config-service/go-client/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Feature represents a configuration feature that can be bound
type Feature = internal.Feature

// PropertyResolverFunc resolves dynamic variation property values
type PropertyResolverFunc = internal.PropertyResolverFunc

// Config holds the configuration for the ConfigClient
type Config struct {
	// Url of the backend configuration gRPC service
	Url string
	// Name:Version of the services to fetch configuration for
	Services map[string]int
	// Interval at which to poll for configuration updates
	PollingInterval time.Duration
	// Interval at which to cleanup unused configuration snapshots
	SnapshotCleanupInterval time.Duration
	// Time after which unused configuration snapshots are deleted
	UnusedSnapshotExpiration time.Duration
}

// options holds internal configuration options
type options struct {
	staticVariation           map[string]string
	dynamicVariationResolvers map[string]PropertyResolverFunc
	features                  []Feature
	changesetOverrider        func(ctx context.Context) *uint32
	loggerFunc                func(ctx context.Context, level slog.Level, msg string, fields ...any)
	productionMode            bool
	fallbackFileLocation      string
	overrides                 internal.Overrides
}

// Option configures the ConfigClient
type Option func(*options)

// WithStaticVariation sets a static variation property value
func WithStaticVariation(property string, value string) Option {
	return func(opts *options) {
		opts.staticVariation[property] = value
	}
}

// WithDynamicVariationResolver sets a dynamic variation property resolver
func WithDynamicVariationResolver(property string, resolver PropertyResolverFunc) Option {
	return func(opts *options) {
		opts.dynamicVariationResolvers[property] = resolver
	}
}

// WithFeatures registers features that can be bound
func WithFeatures(features ...Feature) Option {
	return func(opts *options) {
		opts.features = append(opts.features, features...)
	}
}

// WithChangesetOverrider sets a changeset override function. This is useful for setting a specific (even Open) changeset for development.
func WithChangesetOverrider(overrider func(ctx context.Context) *uint32) Option {
	return func(opts *options) {
		opts.changesetOverrider = overrider
	}
}

// WithLogging sets a custom logging function
func WithLogging(loggerFunc func(ctx context.Context, level slog.Level, msg string, fields ...any)) Option {
	return func(opts *options) {
		opts.loggerFunc = loggerFunc
	}
}

// WithProductionMode enables or disables production mode. In production mode, only services that are published and changesets that are applied can be used.
func WithProductionMode(productionMode bool) Option {
	return func(opts *options) {
		opts.productionMode = productionMode
	}
}

// WithFallbackFileLocation sets the location (base directory) of the fallback file. The latest configuration is stored in a json file and can be loaded from there in case of a service outage.
func WithFallbackFileLocation(fallbackFileLocation string) Option {
	return func(opts *options) {
		opts.fallbackFileLocation = fallbackFileLocation
	}
}

// WithOverride sets an override value. Used to locally override a configuration value in development.
func WithOverride(feature string, key string, value any) Option {
	return func(opts *options) {
		opts.overrides.Set(feature, key, value)
	}
}

// ConfigClient provides access to configuration data
type ConfigClient struct {
	config                  *internal.Config
	registeredFeatureTypes  map[reflect.Type]bool
	snapshotManager         *internal.ConfigurationSnapshotManager
	variationHierarchyStore *internal.VariationHierarchyStore
	grpcConn                *grpc.ClientConn
}

// New creates a new ConfigClient instance
func New(c Config, o ...Option) *ConfigClient {
	services := make([]string, 0, len(c.Services))
	for service, version := range c.Services {
		services = append(services, fmt.Sprintf("%s:%d", service, version))
	}

	opts := &options{
		productionMode:            true,
		staticVariation:           make(map[string]string),
		dynamicVariationResolvers: make(map[string]PropertyResolverFunc),
		overrides:                 make(internal.Overrides),
	}

	for _, o := range o {
		o(opts)
	}

	config := &internal.Config{
		Url:                       c.Url,
		Services:                  services,
		StaticVariation:           opts.staticVariation,
		DynamicVariationResolvers: opts.dynamicVariationResolvers,
		Features:                  opts.features,
		ProductionMode:            opts.productionMode,
		ChangesetOverrider:        opts.changesetOverrider,
		PollingInterval:           c.PollingInterval,
		SnapshotCleanupInterval:   c.SnapshotCleanupInterval,
		UnusedSnapshotExpiration:  c.UnusedSnapshotExpiration,
		FallbackFileLocation:      opts.fallbackFileLocation,
		Logger:                    internal.NewLogger(opts.loggerFunc),
		Overrides:                 opts.overrides,
	}

	if config.PollingInterval == 0 {
		config.PollingInterval = 2 * time.Minute
	}

	if config.SnapshotCleanupInterval == 0 {
		config.SnapshotCleanupInterval = 10 * time.Minute
	}

	if config.UnusedSnapshotExpiration == 0 {
		config.UnusedSnapshotExpiration = 10 * time.Minute
	}

	if opts.loggerFunc != nil {
		config.Logger = internal.NewLogger(opts.loggerFunc)
	}

	registeredFeatureTypes := make(map[reflect.Type]bool)
	for _, feature := range opts.features {
		registeredFeatureTypes[reflect.TypeOf(feature)] = true
	}

	client := &ConfigClient{
		config:                 config,
		registeredFeatureTypes: registeredFeatureTypes,
	}

	return client
}

// Start initializes the client and begins polling for configuration updates
func (c *ConfigClient) Start(ctx context.Context) error {
	if c.config.Url == "" {
		return fmt.Errorf("config service url is not set")
	}

	if c.config.ProductionMode && c.config.ChangesetOverrider != nil {
		return fmt.Errorf("ChangesetOverrider is not supported in production mode")
	}

	if c.grpcConn != nil {
		return fmt.Errorf("config client already started")
	}

	conn, err := grpc.NewClient(
		c.config.Url,
		// TODO: add TLS support
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	c.grpcConn = conn
	configClient := grpcgen.NewConfigServiceClient(conn)
	dataLoader := internal.NewConfigurationDataLoader(configClient, c.config)
	variationHierarchyStore := internal.NewVariationHierarchyStore(dataLoader, c.config)
	pollJob := internal.NewConfigurationPollJob(c.config, variationHierarchyStore, dataLoader)
	snapshotManager := internal.NewConfigurationSnapshotManager(dataLoader, c.config, pollJob)

	c.variationHierarchyStore = variationHierarchyStore
	c.snapshotManager = snapshotManager

	// TODO: start this asynchronously to reduce startup time, and wait in BindFeature
	if err := variationHierarchyStore.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize variation hierarchy: %w", err)
	}

	if err := snapshotManager.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the client
func (c *ConfigClient) Stop(ctx context.Context) error {
	err := c.snapshotManager.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown client: %w", err)
	}

	if c.grpcConn != nil {
		return c.grpcConn.Close()
	}

	return nil
}

// BindFeature binds configuration values to a feature struct
func (c *ConfigClient) BindFeature(ctx context.Context, out Feature) error {
	snapshot, err := c.snapshotManager.GetSnapshot(ctx)
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	variationHierarchy, err := c.variationHierarchyStore.GetVariationHierarchy(ctx)
	if err != nil {
		return fmt.Errorf("failed to get variation hierarchy: %w", err)
	}

	if !c.registeredFeatureTypes[reflect.TypeOf(out)] {
		return fmt.Errorf("feature type %T was not registered in Config.Features", out)
	}

	variationWithParents := make(map[string][]string)
	for property, resolver := range c.config.DynamicVariationResolvers {
		value, err := resolver(ctx)
		if err != nil {
			return fmt.Errorf("failed to get value for property %s: %w", property, err)
		}

		parents, err := variationHierarchy.GetParents(property, value)
		if err != nil {
			return fmt.Errorf("failed resolve property %s with value %s: %w", property, value, err)
		}

		variationWithParents[property] = parents
	}

	return snapshot.BindFeature(out, variationWithParents, c.config.Overrides)
}
