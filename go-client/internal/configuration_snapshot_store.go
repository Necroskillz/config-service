package internal

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"
)

type ConfigurationSnapshotStore struct {
	dataLoader              *ConfigurationDataLoader
	config                  *Config
	variationHierarchyStore *VariationHierarchyStore
	snapshotsMutex          sync.RWMutex
	snapshots               map[uint32]*ConfigurationSnapshot
	lastChangesetID         *uint32
	invalidChangesetIDs     []uint32
	pollCancel              context.CancelFunc
	cleanupCancel           context.CancelFunc
}

func NewConfigurationSnapshotStore(dataLoader *ConfigurationDataLoader, config *Config, variationHierarchyStore *VariationHierarchyStore) *ConfigurationSnapshotStore {
	return &ConfigurationSnapshotStore{
		dataLoader:              dataLoader,
		config:                  config,
		variationHierarchyStore: variationHierarchyStore,
		snapshots:               make(map[uint32]*ConfigurationSnapshot),
	}
}

func (c *ConfigurationSnapshotStore) getSnapshotFileName() string {
	fileName := strings.Builder{}
	fileName.WriteString("configuration")

	for _, property := range slices.Sorted(maps.Keys(c.config.StaticVariation)) {
		value := c.config.StaticVariation[property]
		fileName.WriteString(fmt.Sprintf("_%s-%s", property, value))
	}

	fileName.WriteString(".json")

	return fileName.String()
}

func (c *ConfigurationSnapshotStore) storeSnapshotFile(snapshot *ConfigurationSnapshot) error {
	if err := WriteFallbackFile(c.config, c.getSnapshotFileName(), snapshot); err != nil {
		return fmt.Errorf("failed to write configuration fallback file: %w", err)
	}

	return nil
}

func (c *ConfigurationSnapshotStore) loadSnapshotFile() (*ConfigurationSnapshot, error) {
	snapshot, err := ReadFallbackFile[ConfigurationSnapshot](c.config, c.getSnapshotFileName())
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration fallback file: %w", err)
	}

	return snapshot, nil
}

func (c *ConfigurationSnapshotStore) setSnapshot(ctx context.Context, snapshot *ConfigurationSnapshot) {
	c.snapshotsMutex.Lock()
	defer c.snapshotsMutex.Unlock()

	if len(snapshot.Errors) > 0 {
		c.invalidChangesetIDs = append(c.invalidChangesetIDs, snapshot.ChangesetId)

		c.config.Logger.Error(ctx, "configuration for changeset has errors", "changeset_id", snapshot.ChangesetId, "errors", snapshot.Errors)
	} else {
		c.lastChangesetID = &snapshot.ChangesetId
		c.invalidChangesetIDs = make([]uint32, 0)

		if c.config.IsFallbackFileEnabled() {
			go func() {
				err := c.storeSnapshotFile(snapshot)
				if err != nil {
					c.config.Logger.Error(context.Background(), "failed to store fallback file", "error", err)
				}
			}()
		}
	}

	if len(snapshot.Warnings) > 0 {
		c.config.Logger.Warn(ctx, "configuration for changeset has warnings", "changeset_id", snapshot.ChangesetId, "warnings", snapshot.Warnings)
	}

	c.snapshots[snapshot.ChangesetId] = snapshot
}

func (c *ConfigurationSnapshotStore) getSnapshot(changesetId uint32) (*ConfigurationSnapshot, bool) {
	c.snapshotsMutex.RLock()
	defer c.snapshotsMutex.RUnlock()
	return c.snapshots[changesetId], c.snapshots[changesetId] != nil
}

func (c *ConfigurationSnapshotStore) deleteSnapshot(changesetId uint32) {
	c.snapshotsMutex.Lock()
	defer c.snapshotsMutex.Unlock()
	delete(c.snapshots, changesetId)
}

type ChangesetIDDescriptor struct {
	ChangesetID  *uint32
	IsOverridden bool
}

func (c *ConfigurationSnapshotStore) getLastChangesetID() *uint32 {
	c.snapshotsMutex.RLock()
	defer c.snapshotsMutex.RUnlock()
	return c.lastChangesetID
}

func (c *ConfigurationSnapshotStore) resolveChangesetID(ctx context.Context) ChangesetIDDescriptor {
	descriptor := ChangesetIDDescriptor{
		ChangesetID:  c.lastChangesetID,
		IsOverridden: false,
	}

	if c.config.ChangesetOverrider != nil {
		changesetID := c.config.ChangesetOverrider(ctx)
		if changesetID == nil {
			return descriptor
		}

		descriptor.ChangesetID = changesetID
		descriptor.IsOverridden = true

		return descriptor
	}

	return descriptor
}

func (c *ConfigurationSnapshotStore) Init(ctx context.Context) error {
	cid := c.resolveChangesetID(ctx)

	initialSnapshot, err := c.dataLoader.GetConfiguration(ctx, cid.ChangesetID)
	if err != nil {
		c.config.Logger.Error(ctx, "failed to get initial configuration, trying to load fallback file", "error", err)

		if c.config.IsFallbackFileEnabled() {
			snapshot, err := c.loadSnapshotFile()
			if err != nil {
				return fmt.Errorf("failed to load fallback file: %w", err)
			}

			initialSnapshot = snapshot
		} else {
			return fmt.Errorf("failed to get initial configuration: %w", err)
		}
	}

	initialSnapshot.Validate(c.config.Features)

	if len(initialSnapshot.Errors) > 0 {
		return fmt.Errorf("initial configuration has errors: %v", initialSnapshot.Errors)
	}

	if len(initialSnapshot.Warnings) > 0 {
		c.config.Logger.Warn(ctx, "initial configuration has warnings", "warnings", initialSnapshot.Warnings)
	}

	// If changeset ID is set with an overrider, we don't store it, because it may be an open changeset.
	// And since we don't store it, it doesnt make sense to start polling for next changesets, because we don't have a natural starting point of the last applied changeset.
	// It's possible with this override func approach that the override is there at startup, and then removed later. In that case
	// we start the jobs when the non overriden changeset is requested.
	// TODO: consider if we can reuse (cache) the loaded overriden changeset for some time (e.g. for a duration of a request in case of a web app)
	// TODO: maybe add changeset status so we can store it if it's applied even when overridden
	if !cid.IsOverridden {
		c.setSnapshot(ctx, initialSnapshot)
		c.ensureJobsRunning(ctx)
	}

	return nil
}

func (c *ConfigurationSnapshotStore) Shutdown() {
	c.stopPolling()
	c.stopCleanup()
}

func (c *ConfigurationSnapshotStore) GetSnapshot(ctx context.Context) (*ConfigurationSnapshot, error) {
	cid := c.resolveChangesetID(ctx)

	if cid.ChangesetID != nil {
		snapshot, ok := c.getSnapshot(*cid.ChangesetID)
		if ok {
			return snapshot, nil
		}
	}

	snapshot, err := c.dataLoader.GetConfiguration(ctx, cid.ChangesetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	snapshot.Validate(c.config.Features)

	if !cid.IsOverridden {
		c.setSnapshot(ctx, snapshot)
		c.ensureJobsRunning(ctx)
	} else {
		if len(snapshot.Errors) > 0 {
			return nil, fmt.Errorf("configuration for changeset %d has errors: %v", snapshot.ChangesetId, snapshot.Errors)
		}

		if len(snapshot.Warnings) > 0 {
			c.config.Logger.Warn(ctx, "configuration for changeset has warnings", "changeset_id", snapshot.ChangesetId, "warnings", snapshot.Warnings)
		}
	}

	return snapshot, nil
}

func (c *ConfigurationSnapshotStore) ensureJobsRunning(ctx context.Context) {
	if c.pollCancel == nil {
		c.startPolling(ctx)
	}

	if c.cleanupCancel == nil {
		c.startCleanup(ctx)
	}
}

func (c *ConfigurationSnapshotStore) startPolling(ctx context.Context) error {
	if c.getLastChangesetID() == nil {
		return fmt.Errorf("configuration wasn't loaded. Make sure to call Start() before calling poll()")
	}

	pollCtx, cancel := context.WithCancel(ctx)
	c.pollCancel = cancel
	go c.pollJob(pollCtx)

	return nil
}

func (c *ConfigurationSnapshotStore) stopPolling() {
	if c.pollCancel != nil {
		c.pollCancel()
	}
}

func (c *ConfigurationSnapshotStore) pollJob(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.config.PollingInterval):
			c.config.Logger.Debug(ctx, "poll: starting poll")

			res, err := c.dataLoader.GetNextChangesets(ctx, c.config.Services, *c.getLastChangesetID())
			if err != nil {
				c.config.Logger.Error(ctx, "poll: failed to get next changesets", "error", err)
				continue
			}

			c.config.Logger.Debug(ctx, "poll: next changesets", "changesets", res)

			if len(res) > 0 {
				// Properties/values may have been added since last changeset
				variationHierarchy, err := c.variationHierarchyStore.Refresh(ctx)
				if err != nil {
					// Log error and continue with the hierarchy we already have
					c.config.Logger.Error(ctx, "poll: failed to refresh variation hierarchy", "error", err)
				}

				c.config.Logger.Debug(ctx, "poll: variation hierarchy refreshed", "variation_hierarchy", variationHierarchy)
			}

			for _, changesetId := range slices.Backward(res) {
				snapshot, err := c.dataLoader.GetConfiguration(ctx, &changesetId)
				if err != nil {
					c.config.Logger.Error(ctx, "poll: failed to get configuration for changeset", "changeset_id", changesetId, "error", err)
					continue
				}

				// TODO: if invalid, should we try to load the changeset over and over, or start from the last one even if it has errors?
				c.setSnapshot(ctx, snapshot)

				c.config.Logger.Debug(ctx, "poll: loaded configuration for changeset", "changeset_id", changesetId)
			}
		}
	}
}

func (c *ConfigurationSnapshotStore) startCleanup(ctx context.Context) error {
	cleanupCtx, cancel := context.WithCancel(ctx)
	c.cleanupCancel = cancel
	go c.cleanupJob(cleanupCtx)

	return nil
}

func (c *ConfigurationSnapshotStore) stopCleanup() {
	if c.cleanupCancel != nil {
		c.cleanupCancel()
	}
}

func (c *ConfigurationSnapshotStore) cleanupJob(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.config.SnapshotCleanupInterval):
			c.config.Logger.Debug(ctx, "cleanupOldSnapshots: starting cleanup")
			for changesetId, snapshot := range c.snapshots {
				if snapshot.ChangesetId != *c.getLastChangesetID() && snapshot.LastUsed.Before(time.Now().Add(-c.config.UnusedSnapshotExpiration)) {
					c.deleteSnapshot(changesetId)
					c.config.Logger.Debug(ctx, "cleanupOldSnapshots: deleted snapshot", "changeset_id", changesetId)
				}
			}
			c.config.Logger.Debug(ctx, "cleanupOldSnapshots: cleanup completed")
		}
	}
}
