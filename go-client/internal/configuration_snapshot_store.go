package internal

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"
)

type ConfigurationSnapshotStore struct {
	config            *Config
	snapshotsMutex    sync.RWMutex
	snapshots         map[uint32]*ConfigurationSnapshot
	lastUsed          map[uint32]time.Time
	latestChangesetID *uint32
}

func NewConfigurationSnapshotStore(config *Config) *ConfigurationSnapshotStore {
	return &ConfigurationSnapshotStore{
		config:    config,
		snapshots: make(map[uint32]*ConfigurationSnapshot),
		lastUsed:  make(map[uint32]time.Time),
	}
}

type SetSnapshotOptions struct {
	UpdateLatestChangesetID bool
	Snapshot                *ConfigurationSnapshot
}

func (c *ConfigurationSnapshotStore) Set(options SetSnapshotOptions) {
	c.snapshotsMutex.Lock()
	defer c.snapshotsMutex.Unlock()

	if options.UpdateLatestChangesetID && len(options.Snapshot.Errors) == 0 {
		c.latestChangesetID = &options.Snapshot.ChangesetId
	}

	c.snapshots[options.Snapshot.ChangesetId] = options.Snapshot
}

func (c *ConfigurationSnapshotStore) GetLastChangesetID() *uint32 {
	c.snapshotsMutex.RLock()
	defer c.snapshotsMutex.RUnlock()

	return c.latestChangesetID
}

func (c *ConfigurationSnapshotStore) Get(changesetId uint32) (*ConfigurationSnapshot, bool) {
	c.snapshotsMutex.Lock() // Write lock for usage update
	defer c.snapshotsMutex.Unlock()

	snapshot, exists := c.snapshots[changesetId]
	if exists {
		c.lastUsed[changesetId] = time.Now()
	}

	return snapshot, exists
}

func (c *ConfigurationSnapshotStore) GetLatest() (*ConfigurationSnapshot, error) {
	c.snapshotsMutex.Lock() // Write lock for usage update
	defer c.snapshotsMutex.Unlock()

	if c.latestChangesetID == nil {
		return nil, fmt.Errorf("not initialized")
	}

	snapshot, exists := c.snapshots[*c.latestChangesetID]

	if !exists {
		return nil, fmt.Errorf("latest configuration not found in the store. this should not happen")
	}

	c.lastUsed[*c.latestChangesetID] = time.Now()

	return snapshot, nil
}

func (c *ConfigurationSnapshotStore) CleanupUnused(ctx context.Context) {
	c.snapshotsMutex.Lock()
	defer c.snapshotsMutex.Unlock()

	for changesetId, snapshot := range c.snapshots {
		if c.latestChangesetID == nil || snapshot.ChangesetId != *c.latestChangesetID {
			lastUsedTime, used := c.lastUsed[changesetId]

			if !used || lastUsedTime.Before(time.Now().Add(-c.config.UnusedSnapshotExpiration)) {
				delete(c.snapshots, changesetId)

				c.config.Logger.Debug(ctx, "cleanupOldSnapshots: deleted snapshot", "changeset_id", changesetId)
			}
		}
	}
}

func (c *ConfigurationSnapshotStore) GetMetadata() (map[uint32]*ConfigurationSnapshot, map[uint32]time.Time, *uint32) {
	c.snapshotsMutex.RLock()
	defer c.snapshotsMutex.RUnlock()

	return maps.Clone(c.snapshots), maps.Clone(c.lastUsed), c.latestChangesetID
}
