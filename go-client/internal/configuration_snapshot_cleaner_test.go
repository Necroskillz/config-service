package internal

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/necroskillz/config-service/go-client/internal/test"
	"gotest.tools/v3/assert"
)

type fakeCleanableSnapshotStore struct {
	cleanupCount int
}

func (f *fakeCleanableSnapshotStore) CleanupUnused(ctx context.Context) {
	f.cleanupCount++
}

func TestConfigurationSnapshotCleaner(t *testing.T) {
	type testFixture struct {
		cleaner   *ConfigurationSnapshotCleaner
		fakeStore *fakeCleanableSnapshotStore
		config    *Config
		ctx       context.Context
	}

	setup := func(t *testing.T) *testFixture {
		ctx := context.Background()
		logger := test.NewTestLogger(t)

		config := &Config{
			Logger:                   NewLogger(logger.LogFn),
			SnapshotCleanupInterval:  10 * time.Second,
			UnusedSnapshotExpiration: 20 * time.Second,
		}

		store := &fakeCleanableSnapshotStore{}

		cleaner := NewConfigurationSnapshotCleaner(config, store)

		return &testFixture{cleaner: cleaner, fakeStore: store, config: config, ctx: ctx}
	}

	t.Run("Start stop", func(t *testing.T) {
		fixture := setup(t)

		err := fixture.cleaner.Start(fixture.ctx)
		assert.NilError(t, err)
		assert.Assert(t, fixture.cleaner.IsRunning())

		fixture.cleaner.Stop()
		<-fixture.cleaner.Done()

		assert.Assert(t, !fixture.cleaner.IsRunning())
	})

	t.Run("Returns error when already running", func(t *testing.T) {
		fixture := setup(t)

		err := fixture.cleaner.Start(fixture.ctx)
		assert.NilError(t, err)

		err = fixture.cleaner.Start(fixture.ctx)
		assert.Error(t, err, "cleaner is already running")
	})

	t.Run("Perodically cleans up snapshots unused for more than the expiration period", func(t *testing.T) {
		synctest.Run(func() {
			fixture := setup(t)

			err := fixture.cleaner.Start(fixture.ctx)
			assert.NilError(t, err)

			assert.Equal(t, fixture.fakeStore.cleanupCount, 0)

			time.Sleep(10 * time.Second)
			synctest.Wait()
			assert.Equal(t, fixture.fakeStore.cleanupCount, 1)

			time.Sleep(10 * time.Second)
			synctest.Wait()
			assert.Equal(t, fixture.fakeStore.cleanupCount, 2)

			fixture.cleaner.Stop()
			<-fixture.cleaner.Done()
		})
	})
}
