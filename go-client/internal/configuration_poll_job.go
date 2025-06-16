package internal

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"
	"time"
)

var _ ConfigurationPoller = (*ConfigurationPollJob)(nil)

type ConfigurationPoller interface {
	Start(ctx context.Context, changesetID uint32) error
	IsRunning() bool
	Stop()
	Done() <-chan struct{}
	Snapshots() <-chan *ConfigurationSnapshot
}

type VariationHierarchyRefresher interface {
	Refresh(ctx context.Context) error
}

type ConfigurationPollJob struct {
	config                      *Config
	variationHierarchyRefresher VariationHierarchyRefresher
	dataLoader                  ConfigurationDataLoader
	cancel                      context.CancelFunc
	lastChangesetID             uint32
	running                     atomic.Bool
	done                        chan struct{}
	snapshots                   chan *ConfigurationSnapshot
}

func NewConfigurationPollJob(config *Config, variationHierarchyRefresher VariationHierarchyRefresher, dataLoader ConfigurationDataLoader) *ConfigurationPollJob {
	return &ConfigurationPollJob{
		config:                      config,
		variationHierarchyRefresher: variationHierarchyRefresher,
		dataLoader:                  dataLoader,
		snapshots:                   make(chan *ConfigurationSnapshot),
		done:                        make(chan struct{}),
	}
}

func (c *ConfigurationPollJob) Start(ctx context.Context, changesetID uint32) error {
	if c.IsRunning() {
		return errors.New("poller is already running")
	}

	c.lastChangesetID = changesetID
	c.running.Store(true)

	pollCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	go c.pollJob(pollCtx)

	return nil
}

func (c *ConfigurationPollJob) IsRunning() bool {
	return c.running.Load()
}

func (c *ConfigurationPollJob) Stop() {
	if c.IsRunning() {
		c.cancel()
	}
}

func (c *ConfigurationPollJob) Done() <-chan struct{} {
	return c.done
}

func (c *ConfigurationPollJob) Snapshots() <-chan *ConfigurationSnapshot {
	return c.snapshots
}

func (c *ConfigurationPollJob) poll(ctx context.Context) {
	res, err := c.dataLoader.GetNextChangesets(ctx, c.lastChangesetID)
	if err != nil {
		c.config.Logger.Error(ctx, "poll: failed to get next changesets", "error", err)
		return
	}

	c.config.Logger.Debug(ctx, "poll: next changesets", "changesets", res)

	if len(res) > 0 {
		err := c.variationHierarchyRefresher.Refresh(ctx)
		if err != nil {
			c.config.Logger.Error(ctx, "poll: failed to refresh variation hierarchy", "error", err)
		} else {
			c.config.Logger.Debug(ctx, "poll: variation hierarchy refreshed")
		}

	}

	changesetIDUpdated := false

	for _, changesetId := range slices.Backward(res) {
		snapshot, err := c.dataLoader.GetConfiguration(ctx, &changesetId)
		if err != nil {
			c.config.Logger.Error(ctx, "poll: failed to get configuration for changeset", "changeset_id", changesetId, "error", err)
			continue
		}

		c.config.Logger.Debug(ctx, "poll: loaded configuration for changeset", "changeset_id", changesetId)

		if !changesetIDUpdated {
			c.lastChangesetID = changesetId
			changesetIDUpdated = true
		}

		c.snapshots <- snapshot

		if len(snapshot.Errors) == 0 {
			break
		}
	}
}

func (c *ConfigurationPollJob) pollJob(ctx context.Context) {
	defer c.running.Store(false)
	defer close(c.done)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.config.PollingInterval):
			c.config.Logger.Debug(ctx, "poll: starting poll")

			c.poll(ctx)
		}
	}
}
