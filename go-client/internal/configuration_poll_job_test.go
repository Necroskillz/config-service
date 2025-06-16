package internal

import (
	"context"
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/necroskillz/config-service/go-client/internal/test"
	"github.com/stretchr/testify/mock"
	"gotest.tools/v3/assert"
)

func TestConfigurationPollJob(t *testing.T) {
	TestSnapshot := func(changesetID uint32, errors ...string) *ConfigurationSnapshot {
		return &ConfigurationSnapshot{
			ChangesetId: changesetID,
			Errors:      errors,
		}
	}

	type testFixture struct {
		pollJob                     *ConfigurationPollJob
		config                      *Config
		ctx                         context.Context
		variationHierarchyRefresher *MockVariationHierarchyRefresher
		dataLoaderMock              *MockConfigurationDataLoader
		logger                      *test.TestLogger
	}

	setup := func(t *testing.T) *testFixture {
		ctx := context.Background()
		logger := test.NewTestLogger(t)
		variationHierarchyRefresher := &MockVariationHierarchyRefresher{}
		dataLoaderMock := NewMockConfigurationDataLoader(t)

		config := &Config{
			Logger:          NewLogger(logger.LogFn),
			PollingInterval: 10 * time.Second,
			Services:        []string{"service1", "service2"},
			StaticVariation: map[string]string{
				"env":    "dev",
				"domain": "example.com",
			},
		}

		pollJob := NewConfigurationPollJob(
			config,
			variationHierarchyRefresher,
			dataLoaderMock,
		)

		return &testFixture{config: config, pollJob: pollJob, ctx: ctx, variationHierarchyRefresher: variationHierarchyRefresher, dataLoaderMock: dataLoaderMock, logger: logger}
	}

	t.Run("Start stop", func(t *testing.T) {
		fixture := setup(t)

		err := fixture.pollJob.Start(fixture.ctx, 1)
		assert.NilError(t, err)
		assert.Assert(t, fixture.pollJob.IsRunning())

		fixture.pollJob.Stop()
		<-fixture.pollJob.Done()

		assert.Assert(t, !fixture.pollJob.IsRunning())
	})

	t.Run("Returns error when already running", func(t *testing.T) {
		fixture := setup(t)

		err := fixture.pollJob.Start(fixture.ctx, 1)
		assert.NilError(t, err)

		err = fixture.pollJob.Start(fixture.ctx, 1)
		assert.Error(t, err, "poller is already running")

		fixture.pollJob.Stop()
		<-fixture.pollJob.Done()
	})

	t.Run("Poll", func(t *testing.T) {
		type testCase struct {
			nextChangesets                 []uint32
			variationHierarchyRefreshError error
			configurationLoadErrors        []error
			nextChangesetsError            error
			nextConfigurations             []*ConfigurationSnapshot
			nextPollChangesetID            uint32
		}

		run := func(t *testing.T, testCase testCase) {
			synctest.Run(func() {
				fixture := setup(t)

				fixture.dataLoaderMock.EXPECT().GetNextChangesets(mock.Anything, uint32(1)).Return(testCase.nextChangesets, testCase.nextChangesetsError)
				fixture.dataLoaderMock.EXPECT().GetNextChangesets(mock.Anything, testCase.nextPollChangesetID).Return([]uint32{}, nil)

				if testCase.nextChangesetsError == nil {
					fixture.variationHierarchyRefresher.EXPECT().Refresh(mock.Anything).Return(testCase.variationHierarchyRefreshError)

					for i, changesetID := range slices.Backward(testCase.nextChangesets) {
						var err error
						if testCase.configurationLoadErrors != nil {
							err = testCase.configurationLoadErrors[i]
						}

						snapshot := testCase.nextConfigurations[i]

						fixture.dataLoaderMock.EXPECT().GetConfiguration(mock.Anything, &changesetID).Return(snapshot, err)

						if snapshot != nil && len(snapshot.Errors) == 0 {
							break
						}
					}
				}

				fixture.pollJob.Start(fixture.ctx, 1)

				time.Sleep(10 * time.Second)
				synctest.Wait()

				for i, snapshot := range slices.Backward(testCase.nextConfigurations) {
					if testCase.configurationLoadErrors != nil && testCase.configurationLoadErrors[i] != nil {
						continue
					}

					publishedSnapshot := <-fixture.pollJob.Snapshots()
					assert.DeepEqual(t, publishedSnapshot, snapshot)

					if len(snapshot.Errors) == 0 {
						break
					}
				}

				if testCase.nextChangesetsError != nil {
					fixture.logger.AssertLog(t, test.WithMessage("poll: failed to get next changesets"))
				}

				if testCase.variationHierarchyRefreshError != nil {
					fixture.logger.AssertLog(t, test.WithMessage("poll: failed to refresh variation hierarchy"))
				}

				if testCase.configurationLoadErrors != nil {
					for range testCase.configurationLoadErrors {
						fixture.logger.AssertLog(t, test.WithMessage("poll: failed to get configuration for changeset"))
					}
				}

				time.Sleep(10 * time.Second)
				synctest.Wait()

				fixture.pollJob.Stop()
				<-fixture.pollJob.Done()
			})
		}

		cases := map[string]testCase{
			"normal operation": {
				nextChangesets: []uint32{2, 3},
				nextConfigurations: []*ConfigurationSnapshot{
					TestSnapshot(2),
					TestSnapshot(3),
				},
				nextPollChangesetID: 3,
			},
			"error getting next changesets": {
				nextChangesets:      []uint32{},
				nextChangesetsError: errors.New("error getting next changesets"),
				nextPollChangesetID: 1,
			},
			"error refreshing variation hierarchy": {
				nextChangesets:                 []uint32{2, 3},
				variationHierarchyRefreshError: errors.New("error refreshing variation hierarchy"),
				nextConfigurations: []*ConfigurationSnapshot{
					TestSnapshot(2),
					TestSnapshot(3),
				},
				nextPollChangesetID: 3,
			},
			"error loading configuration": {
				nextChangesets:          []uint32{2, 3},
				configurationLoadErrors: []error{nil, errors.New("error loading configuration")},
				nextConfigurations: []*ConfigurationSnapshot{
					TestSnapshot(2),
					nil,
				},
				nextPollChangesetID: 2,
			},
			"invalid configuration": {
				nextChangesets: []uint32{2, 3},
				nextConfigurations: []*ConfigurationSnapshot{
					TestSnapshot(2),
					TestSnapshot(3, "error"),
				},
				nextPollChangesetID: 3,
			},
		}

		test.RunCases(t, run, cases)
	})
}
