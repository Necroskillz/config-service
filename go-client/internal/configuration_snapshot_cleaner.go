package internal

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

type ConfigurationSnapshotCleaner struct {
	config  *Config
	store   CleanableSnapshotStore
	running atomic.Bool
	cancel  context.CancelFunc
	done    chan struct{}
}

type CleanableSnapshotStore interface {
	CleanupUnused(ctx context.Context)
}

func NewConfigurationSnapshotCleaner(config *Config, store CleanableSnapshotStore) *ConfigurationSnapshotCleaner {
	return &ConfigurationSnapshotCleaner{
		config: config,
		store:  store,
		done:   make(chan struct{}),
	}
}

func (c *ConfigurationSnapshotCleaner) Start(ctx context.Context) error {
	if c.IsRunning() {
		return errors.New("cleaner is already running")
	}

	cleanupCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.running.Store(true)
	go c.cleanupJob(cleanupCtx)

	return nil
}

func (c *ConfigurationSnapshotCleaner) IsRunning() bool {
	return c.running.Load()
}

func (c *ConfigurationSnapshotCleaner) Stop() {
	if c.IsRunning() {
		c.cancel()
	}
}

func (c *ConfigurationSnapshotCleaner) Done() <-chan struct{} {
	return c.done
}

func (c *ConfigurationSnapshotCleaner) cleanupJob(ctx context.Context) {
	defer c.running.Store(false)
	defer close(c.done)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.config.SnapshotCleanupInterval):
			c.config.Logger.Debug(ctx, "cleanupOldSnapshots: starting cleanup")

			c.store.CleanupUnused(ctx)

			c.config.Logger.Debug(ctx, "cleanupOldSnapshots: cleanup completed")
		}
	}
}
