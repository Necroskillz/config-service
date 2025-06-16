package internal

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/necroskillz/config-service/go-client/internal/test"
	"gotest.tools/v3/assert"
)

func TestConfigurationSnapshotStore(t *testing.T) {
	TestSnapshot := func() *ConfigurationSnapshot {
		return &ConfigurationSnapshot{
			ChangesetId: 1,
		}
	}

	type testFixture struct {
		store  *ConfigurationSnapshotStore
		config *Config
		ctx    context.Context
	}

	setup := func(t *testing.T) *testFixture {
		logger := test.NewTestLogger(t)

		config := &Config{
			Logger:                   NewLogger(logger.LogFn),
			UnusedSnapshotExpiration: 10 * time.Second,
		}

		return &testFixture{
			config: config,
			store:  NewConfigurationSnapshotStore(config),
		}
	}

	t.Run("Get", func(t *testing.T) {
		t.Run("Retrieves stored snapshot", func(t *testing.T) {
			fixture := setup(t)

			testSnapshot := TestSnapshot()

			fixture.store.Set(SetSnapshotOptions{
				Snapshot: testSnapshot,
			})

			snapshot, exists := fixture.store.Get(testSnapshot.ChangesetId)
			assert.Assert(t, exists)
			assert.Equal(t, snapshot, testSnapshot)
		})

		t.Run("Returns false if snapshot is not found", func(t *testing.T) {
			fixture := setup(t)
			testSnapshot := TestSnapshot()

			fixture.store.Set(SetSnapshotOptions{
				Snapshot: testSnapshot,
			})

			snapshot, exists := fixture.store.Get(2)
			assert.Assert(t, !exists)
			assert.Assert(t, snapshot == nil)
		})
	})

	t.Run("GetLatest", func(t *testing.T) {
		t.Run("Sets stored snapshot as latest", func(t *testing.T) {
			fixture := setup(t)
			testSnapshot := TestSnapshot()

			fixture.store.Set(SetSnapshotOptions{
				Snapshot:                testSnapshot,
				UpdateLatestChangesetID: true,
			})

			latestChangesetID := *fixture.store.GetLastChangesetID()
			assert.Equal(t, latestChangesetID, testSnapshot.ChangesetId)

			snapshot, err := fixture.store.GetLatest()
			assert.NilError(t, err)
			assert.Equal(t, snapshot, testSnapshot)
		})

		t.Run("Does not set stored snapshot as latest if it has errors", func(t *testing.T) {
			fixture := setup(t)
			validSnapshot := TestSnapshot()
			fixture.store.Set(SetSnapshotOptions{
				Snapshot:                validSnapshot,
				UpdateLatestChangesetID: true,
			})

			errorSnapshot := TestSnapshot()
			errorSnapshot.Errors = []string{"error"}
			errorSnapshot.ChangesetId = 2

			fixture.store.Set(SetSnapshotOptions{
				Snapshot:                errorSnapshot,
				UpdateLatestChangesetID: true,
			})

			latestChangesetID := *fixture.store.GetLastChangesetID()
			assert.Equal(t, latestChangesetID, validSnapshot.ChangesetId)

			snapshot, err := fixture.store.GetLatest()
			assert.NilError(t, err)
			assert.Equal(t, snapshot, validSnapshot)
		})

		t.Run("Returns unintialized error if no snapshots are stored", func(t *testing.T) {
			fixture := setup(t)

			latestChangesetID := fixture.store.GetLastChangesetID()
			assert.Assert(t, latestChangesetID == nil)

			_, err := fixture.store.GetLatest()
			assert.Error(t, err, "not initialized")
		})

		t.Run("Returns unintialized error if only invalid snapshots are stored", func(t *testing.T) {
			fixture := setup(t)
			testSnapshot := TestSnapshot()
			testSnapshot.Errors = []string{"error"}

			fixture.store.Set(SetSnapshotOptions{
				Snapshot:                testSnapshot,
				UpdateLatestChangesetID: true,
			})

			latestChangesetID := fixture.store.GetLastChangesetID()
			assert.Assert(t, latestChangesetID == nil)

			_, err := fixture.store.GetLatest()
			assert.Error(t, err, "not initialized")
		})
	})

	t.Run("GetMetadata", func(t *testing.T) {
		t.Run("Retrieves metadata about stored snapshots", func(t *testing.T) {
			fixture := setup(t)
			testSnapshot := TestSnapshot()

			fixture.store.Set(SetSnapshotOptions{
				Snapshot:                testSnapshot,
				UpdateLatestChangesetID: true,
			})

			snapshots, usage, lastChangesetID := fixture.store.GetMetadata()
			assert.Equal(t, len(snapshots), 1)
			assert.Equal(t, snapshots[testSnapshot.ChangesetId], testSnapshot)
			assert.Equal(t, len(usage), 0)
			assert.Equal(t, *lastChangesetID, testSnapshot.ChangesetId)
		})

		t.Run("Returns nil latest changeset id if no current snapshots are stored", func(t *testing.T) {
			fixture := setup(t)
			testSnapshot := TestSnapshot()

			fixture.store.Set(SetSnapshotOptions{
				Snapshot: testSnapshot,
			})

			snapshots, usage, lastChangesetID := fixture.store.GetMetadata()
			assert.Equal(t, len(snapshots), 1)
			assert.Equal(t, snapshots[testSnapshot.ChangesetId], testSnapshot)
			assert.Equal(t, len(usage), 0)
			assert.Assert(t, lastChangesetID == nil)
		})

		t.Run("Tracks usage when snapshot is retrieved", func(t *testing.T) {
			type testCase struct {
				retrieveFn func(store *ConfigurationSnapshotStore)
			}

			run := func(t *testing.T, test testCase) {
				synctest.Run(func() {
					fixture := setup(t)
					testSnapshot := TestSnapshot()
					now := time.Now()

					fixture.store.Set(SetSnapshotOptions{
						Snapshot:                testSnapshot,
						UpdateLatestChangesetID: true,
					})

					fixture.store.Get(testSnapshot.ChangesetId)

					snapshots, usage, lastChangesetID := fixture.store.GetMetadata()

					assert.Equal(t, len(snapshots), 1)
					assert.Equal(t, snapshots[testSnapshot.ChangesetId], testSnapshot)
					assert.Equal(t, len(usage), 1)
					assert.Equal(t, usage[testSnapshot.ChangesetId], now)
					assert.Equal(t, *lastChangesetID, testSnapshot.ChangesetId)
				})
			}

			cases := map[string]testCase{
				"Get": {
					retrieveFn: func(store *ConfigurationSnapshotStore) {
						store.Get(1)
					},
				},
				"GetLatest": {
					retrieveFn: func(store *ConfigurationSnapshotStore) {
						store.GetLatest()
					},
				},
			}

			test.RunCases(t, run, cases)
		})
	})

	t.Run("CleanupUnused", func(t *testing.T) {
		setSnapshot := func(fixture *testFixture, changesetId uint32) {
			fixture.store.Set(SetSnapshotOptions{
				Snapshot: &ConfigurationSnapshot{ChangesetId: changesetId},
			})
		}

		assertSnapshotExists := func(fixture *testFixture, changesetId uint32) {
			t.Helper()

			snapshot, exists := fixture.store.Get(changesetId)
			assert.Assert(t, exists)
			assert.Assert(t, snapshot != nil)
		}

		assertSnapshotDoesNotExist := func(fixture *testFixture, changesetId uint32) {
			t.Helper()

			snapshot, exists := fixture.store.Get(changesetId)
			assert.Assert(t, !exists)
			assert.Assert(t, snapshot == nil)
		}

		t.Run("Deletes unused snapshots", func(t *testing.T) {
			synctest.Run(func() {
				fixture := setup(t)

				// Add 2 snapshots and use them
				setSnapshot(fixture, 1)
				setSnapshot(fixture, 2)
				fixture.store.Get(1)
				fixture.store.Get(2)

				// Wait until the snapshots are considered unused
				time.Sleep(11 * time.Second)
				synctest.Wait()

				// Add a new snapshot and use it
				setSnapshot(fixture, 3)
				fixture.store.Get(3)

				fixture.store.CleanupUnused(fixture.ctx)

				time.Sleep(10 * time.Second)
				synctest.Wait()

				// First two snapshots should be deleted
				assertSnapshotDoesNotExist(fixture, 1)
				assertSnapshotDoesNotExist(fixture, 2)
				assertSnapshotExists(fixture, 3)
			})
		})
	})

	t.Run("Concurrent reads and writes", func(t *testing.T) {
		fixture := setup(t)

		// Pre-populate with some snapshots
		for i := uint32(1); i <= 10; i++ {
			fixture.store.Set(SetSnapshotOptions{
				Snapshot:                &ConfigurationSnapshot{ChangesetId: i},
				UpdateLatestChangesetID: true,
			})
		}

		const numGoroutines = 50
		const opsPerGoroutine = 100

		doneChan := make(chan struct{}, numGoroutines)

		// Start multiple goroutines doing concurrent operations
		for i := range numGoroutines {
			go func(goroutineID int) {
				defer func() { doneChan <- struct{}{} }()

				for j := range opsPerGoroutine {
					switch j % 4 {
					case 0: // Read operations
						_, _ = fixture.store.Get(uint32((j % 10) + 1))
					case 1: // GetLatest operations
						_, _ = fixture.store.GetLatest()
					case 2: // Write operations
						newID := uint32(100 + goroutineID*opsPerGoroutine + j)
						fixture.store.Set(SetSnapshotOptions{
							Snapshot:                &ConfigurationSnapshot{ChangesetId: newID},
							UpdateLatestChangesetID: true,
						})
					case 3: // Metadata operations
						_, _, _ = fixture.store.GetMetadata()
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for range numGoroutines {
			<-doneChan
		}
	})
}
