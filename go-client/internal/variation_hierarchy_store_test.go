package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"

	"github.com/necroskillz/config-service/go-client/internal/test"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
)

func TestVariationHierarchyStore(t *testing.T) {
	defaultVariationHierarchy := NewVariationHierarchy(test.NewTestVariationHierarchyResponseBuilder().WithValue("env", "dev").Response())

	type testFixture struct {
		logger         *test.TestLogger
		config         *Config
		store          *VariationHierarchyStore
		ctx            context.Context
		dataLoaderMock *MockConfigurationDataLoader
	}

	setup := func(t *testing.T) *testFixture {
		logger := test.NewTestLogger(t)
		config := &Config{
			Logger: NewLogger(logger.LogFn),
		}
		dataLoaderMock := NewMockConfigurationDataLoader(t)

		return &testFixture{
			logger:         logger,
			config:         config,
			store:          NewVariationHierarchyStore(dataLoaderMock, config),
			ctx:            context.Background(),
			dataLoaderMock: dataLoaderMock,
		}
	}

	assertVariationHierarchy := func(t *testing.T, fixture *testFixture, expectedVariationHierarchy *VariationHierarchy) {
		variationHierarchy, err := fixture.store.GetVariationHierarchy(fixture.ctx)
		assert.NilError(t, err)

		assert.DeepEqual(t, variationHierarchy, expectedVariationHierarchy)
	}

	defaultInit := func(t *testing.T, fixture *testFixture) {
		fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(defaultVariationHierarchy, nil)

		err := fixture.store.Init(fixture.ctx)

		assert.NilError(t, err)
	}

	t.Run("Init", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			fixture := setup(t)

			defaultInit(t, fixture)

			assertVariationHierarchy(t, fixture, defaultVariationHierarchy)
		})

		t.Run("Caches network response", func(t *testing.T) {
			fixture := setup(t)

			defaultInit(t, fixture)

			for range 10 {
				assertVariationHierarchy(t, fixture, defaultVariationHierarchy)
			}
		})

		t.Run("Saves fallback file and loads it on init when loading data fails", func(t *testing.T) {
			fixture := setup(t)

			tempDir := fs.NewDir(t, "fallback-test")

			fixture.config.FallbackFileLocation = tempDir.Path()

			defaultInit(t, fixture)

			variationHierarchy, err := fixture.store.GetVariationHierarchy(fixture.ctx)
			assert.NilError(t, err)

			serialized, err := json.Marshal(variationHierarchy)
			assert.NilError(t, err)

			test.WaitForFile(t, filepath.Join(tempDir.Path(), "variation_hierarchy.json"))

			expected := fs.Expected(t, fs.WithFile("variation_hierarchy.json", string(serialized)))
			assert.Assert(t, fs.Equal(tempDir.Path(), expected))

			fixture2 := setup(t)

			fixture2.config.FallbackFileLocation = tempDir.Path()

			fixture2.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture2.ctx).Return(nil, errors.New("error msg"))

			err = fixture2.store.Init(fixture2.ctx)
			assert.NilError(t, err)

			fixture2.logger.AssertLog(t,
				test.WithLevel(slog.LevelError),
				test.WithMessage("failed to get variation hierarchy, trying to load fallback file"),
				test.WithField("error", func(t *testing.T, value any) cmp.Comparison {
					return cmp.ErrorContains(value.(error), "error msg")
				}),
			)

			assertVariationHierarchy(t, fixture2, defaultVariationHierarchy)
		})

		t.Run("Returns error when unable to load fallback file", func(t *testing.T) {
			fixture := setup(t)

			tempDir := fs.NewDir(t, "no-fallback-file-test")

			fixture.config.FallbackFileLocation = tempDir.Path()

			fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(nil, errors.New("error msg"))

			err := fixture.store.Init(fixture.ctx)
			assert.ErrorContains(t, err, "failed to load variation hierarchy fallback file")

			fixture.logger.AssertLog(t,
				test.WithLevel(slog.LevelError),
				test.WithMessage("failed to get variation hierarchy, trying to load fallback file"),
				test.WithField("error", func(t *testing.T, value any) cmp.Comparison {
					return cmp.ErrorContains(value.(error), "error msg")
				}),
			)
		})

		t.Run("Logs error when unable to write fallback file", func(t *testing.T) {
			fixture := setup(t)

			fixture.config.FallbackFileLocation = "invalid\x00path"
			fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(defaultVariationHierarchy, nil)

			defaultInit(t, fixture)

			fixture.logger.AssertLog(t,
				test.WithLevel(slog.LevelError),
				test.WithMessage("failed to store variation hierarchy fallback file"),
				test.WithField("error", func(t *testing.T, value any) cmp.Comparison {
					return cmp.ErrorContains(value.(error), "failed to write variation hierarchy fallback file")
				}),
			)
		})

		t.Run("Returns error when unable to make network request", func(t *testing.T) {
			fixture := setup(t)

			fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(nil, errors.New("error msg"))

			err := fixture.store.Init(fixture.ctx)
			assert.ErrorContains(t, err, "failed to init variation hierarchy")
			assert.ErrorContains(t, err, "error msg")
		})
	})

	t.Run("GetVariationHierarchy", func(t *testing.T) {
		t.Run("Returns error when store is not initialized", func(t *testing.T) {
			fixture := setup(t)

			_, err := fixture.store.GetVariationHierarchy(fixture.ctx)

			assert.Error(t, err, "variation hierarchy store is not initialized")
		})
	})

	t.Run("Refresh", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			fixture := setup(t)

			defaultInit(t, fixture)
			assertVariationHierarchy(t, fixture, defaultVariationHierarchy)

			refreshVariationHierarchy := NewVariationHierarchy(test.NewTestVariationHierarchyResponseBuilder().
				WithValue("env", "qa1").
				Response())

			fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(refreshVariationHierarchy, nil)

			err := fixture.store.Refresh(fixture.ctx)

			assert.NilError(t, err)

			assertVariationHierarchy(t, fixture, defaultVariationHierarchy)
		})

		t.Run("Returns error when loading data fails", func(t *testing.T) {
			fixture := setup(t)

			fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(nil, errors.New("error msg"))

			err := fixture.store.Init(fixture.ctx)

			assert.ErrorContains(t, err, "error msg")
		})

		t.Run("Updates fallback file", func(t *testing.T) {
			fixture := setup(t)

			tempDir := fs.NewDir(t, "refresh-fallback-test")

			fixture.config.FallbackFileLocation = tempDir.Path()

			// Init to create fallback file
			defaultInit(t, fixture)

			variationHierarchy, err := fixture.store.GetVariationHierarchy(fixture.ctx)
			assert.NilError(t, err)

			serialized, err := json.Marshal(variationHierarchy)
			assert.NilError(t, err)

			fallbackFile := filepath.Join(tempDir.Path(), "variation_hierarchy.json")

			// Wait for file and check contents
			test.WaitForFile(t, fallbackFile)

			expected := fs.Expected(t, fs.WithFile("variation_hierarchy.json", string(serialized)))
			assert.Assert(t, fs.Equal(tempDir.Path(), expected))

			// Refresh to update fallback file
			refreshVariationHierarchy := NewVariationHierarchy(test.NewTestVariationHierarchyResponseBuilder().
				WithValue("env", "qa1").
				Response())

			fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(refreshVariationHierarchy, nil)

			err = fixture.store.Refresh(fixture.ctx)
			assert.NilError(t, err)

			refreshedVariationHierarchy, err := fixture.store.GetVariationHierarchy(fixture.ctx)
			assert.NilError(t, err)
			serialized, err = json.Marshal(refreshedVariationHierarchy)
			assert.NilError(t, err)

			// Wait for file update and check contents
			test.WaitForFileUpdate(t, fallbackFile)

			expected = fs.Expected(t, fs.WithFile("variation_hierarchy.json", string(serialized)))
			assert.Assert(t, fs.Equal(tempDir.Path(), expected))
		})
	})

	t.Run("Concurrent access", func(t *testing.T) {
		fixture := setup(t)

		defaultInit(t, fixture)
		var wg sync.WaitGroup
		iterations := 100

		for range iterations {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fixture.store.GetVariationHierarchy(fixture.ctx)
			}()
		}

		for range iterations {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fixture.dataLoaderMock.EXPECT().GetVariationHierarchy(fixture.ctx).Return(defaultVariationHierarchy, nil)

				fixture.store.Refresh(fixture.ctx)
			}()
		}

		wg.Wait()
	})
}
