package internal

import (
	"context"
	"fmt"
)

type ConfigurationSnapshotManager struct {
	config       *Config
	dataLoader   ConfigurationDataLoader
	store        *ConfigurationSnapshotStore
	fallbackFile *ConfigurationFallbackFile
	poller       ConfigurationPoller
	cleaner      *ConfigurationSnapshotCleaner
}

func NewConfigurationSnapshotManager(
	dataLoader ConfigurationDataLoader,
	config *Config,
	poller ConfigurationPoller,
) *ConfigurationSnapshotManager {
	store := NewConfigurationSnapshotStore(config)

	manager := &ConfigurationSnapshotManager{
		dataLoader:   dataLoader,
		config:       config,
		store:        store,
		fallbackFile: NewConfigurationFallbackFile(config),
		poller:       poller,
		cleaner:      NewConfigurationSnapshotCleaner(config, store),
	}

	go func() {
		for snapshot := range poller.Snapshots() {
			manager.storeSnapshot(context.Background(), StoreSnapshotOptions{
				IsOverriden: false,
				Snapshot:    snapshot,
			})
		}
	}()

	return manager
}

func (c *ConfigurationSnapshotManager) Init(ctx context.Context) error {
	_, err := c.loadInitialSnapshot(ctx)

	if err != nil {
		return fmt.Errorf("failed to get initial configuration: %w", err)
	}

	return nil
}

func (c *ConfigurationSnapshotManager) Shutdown(ctx context.Context) error {
	c.poller.Stop()
	c.cleaner.Stop()

	select {
	case <-c.poller.Done():
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-c.cleaner.Done():
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (c *ConfigurationSnapshotManager) GetSnapshot(ctx context.Context) (*ConfigurationSnapshot, error) {
	if c.config.ChangesetOverrider != nil {
		changesetID := c.config.ChangesetOverrider(ctx)
		if changesetID != nil {
			snapshot, ok := c.store.Get(*changesetID)
			if ok {
				if len(snapshot.Errors) > 0 {
					return nil, fmt.Errorf("configuration for changeset %d has errors: %v", snapshot.ChangesetId, snapshot.Errors)
				}

				return snapshot, nil
			} else {
				return c.loadOverriddenSnapshot(ctx, *changesetID)
			}
		}
	}

	// If we start the client with an overridden changeset, and after the initial load the override func
	// starts returning nil, we start up the normal process
	if c.store.GetLastChangesetID() == nil {
		return c.loadInitialSnapshot(ctx)
	}

	snapshot, err := c.store.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest configuration: %w", err)
	}

	return snapshot, nil
}

func (c *ConfigurationSnapshotManager) loadInitialSnapshot(ctx context.Context) (*ConfigurationSnapshot, error) {
	snapshot, err := c.dataLoader.GetConfiguration(ctx, nil)
	if err != nil {
		if c.config.IsFallbackFileEnabled() {
			c.config.Logger.Error(ctx, "failed to get configuration, trying to load fallback file", "error", err)

			fallback, err := c.fallbackFile.Read()
			if err != nil {
				return nil, fmt.Errorf("failed to load fallback file: %w", err)
			}

			snapshot = fallback
		} else {
			return nil, fmt.Errorf("failed to get configuration: %w", err)
		}
	}

	if len(snapshot.Errors) > 0 {
		return nil, fmt.Errorf("configuration for changeset %d has errors: %v", snapshot.ChangesetId, snapshot.Errors)
	}

	c.storeSnapshot(ctx, StoreSnapshotOptions{
		IsOverriden: false,
		Snapshot:    snapshot,
	})
	c.ensureJobsRunning(ctx)

	return snapshot, nil
}

func (c *ConfigurationSnapshotManager) loadOverriddenSnapshot(ctx context.Context, changesetID uint32) (*ConfigurationSnapshot, error) {
	snapshot, err := c.dataLoader.GetConfiguration(ctx, &changesetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	c.storeSnapshot(ctx, StoreSnapshotOptions{
		IsOverriden: true,
		Snapshot:    snapshot,
	})

	return snapshot, nil
}

type StoreSnapshotOptions struct {
	Snapshot    *ConfigurationSnapshot
	IsOverriden bool
}

func (c *ConfigurationSnapshotManager) storeSnapshot(ctx context.Context, options StoreSnapshotOptions) {
	if options.Snapshot.AppliedAt != nil {
		c.store.Set(SetSnapshotOptions{
			Snapshot:                options.Snapshot,
			UpdateLatestChangesetID: !options.IsOverriden,
		})
	}

	if len(options.Snapshot.Errors) > 0 {
		c.config.Logger.Error(ctx, "configuration for changeset has errors", "changeset_id", options.Snapshot.ChangesetId, "errors", options.Snapshot.Errors)
	} else {
		if !options.IsOverriden && c.config.IsFallbackFileEnabled() {
			go func() {
				err := c.fallbackFile.Write(options.Snapshot)
				if err != nil {
					c.config.Logger.Error(context.Background(), "failed to store fallback file", "error", err)
				}
			}()
		}
	}

	if len(options.Snapshot.Warnings) > 0 {
		c.config.Logger.Warn(ctx, "configuration for changeset has warnings", "changeset_id", options.Snapshot.ChangesetId, "warnings", options.Snapshot.Warnings)
	}
}

func (c *ConfigurationSnapshotManager) ensureJobsRunning(ctx context.Context) {
	lastChangesetID := *c.store.GetLastChangesetID()

	if !c.poller.IsRunning() {
		err := c.poller.Start(ctx, lastChangesetID)
		if err != nil {
			c.config.Logger.Error(ctx, "failed to start poller", "error", err)
		} else {
			c.config.Logger.Debug(ctx, "polling job started")
		}
	}

	if !c.cleaner.IsRunning() {
		err := c.cleaner.Start(ctx)
		if err != nil {
			c.config.Logger.Error(ctx, "failed to start cleaner", "error", err)
		} else {
			c.config.Logger.Debug(ctx, "cleanup job started")
		}
	}
}
