package auth_test

import (
	"testing"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/util/ptr"
	"github.com/necroskillz/config-service/util/test"
	"gotest.tools/v3/assert"
)

func TestVariationPermission(t *testing.T) {
	variationPermission := auth.VariationPermission{
		ServiceID: 1,
		FeatureID: 2,
		KeyID:     3,
		Level:     constants.PermissionAdmin,
		PropertyValues: map[uint]string{
			1: "value1",
			2: "value2",
		},
	}

	serviceId := uint(1)
	featureId := uint(2)
	keyId := uint(3)

	t.Run("Match", func(t *testing.T) {
		type testCase struct {
			serviceId          uint
			featureId          *uint
			keyId              *uint
			variation          map[uint][]string
			expectedPermission constants.PermissionLevel
		}

		run := func(t *testing.T, tc testCase) {
			permission := variationPermission.Match(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, permission, tc.expectedPermission)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionAdmin},
			"parent match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value", "value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionAdmin},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionViewer},
			"different feature": {serviceId: serviceId, featureId: ptr.To(uint(4)), keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionViewer},
			"different key": {serviceId: serviceId, featureId: &featureId, keyId: ptr.To(uint(4)), variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionViewer},
			"different property": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				3: {"value2"},
			}, expectedPermission: constants.PermissionViewer},
			"extra property": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
				3: {"value3"},
			}, expectedPermission: constants.PermissionAdmin},
			"nil variation": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"nil key":       {serviceId: serviceId, featureId: &featureId, keyId: nil, variation: nil, expectedPermission: constants.PermissionViewer},
			"nil feature":   {serviceId: serviceId, featureId: nil, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"only service":  {serviceId: serviceId, featureId: nil, keyId: nil, variation: nil, expectedPermission: constants.PermissionViewer},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MatchAny", func(t *testing.T) {
		type testCase struct {
			serviceId     uint
			featureId     *uint
			keyId         *uint
			variation     map[uint][]string
			expectedMatch bool
		}

		run := func(t *testing.T, tc testCase) {
			match := variationPermission.MatchAny(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, match, tc.expectedMatch)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedMatch: true},
			"key match":         {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: true},
			"feature match":     {serviceId: serviceId, featureId: &featureId, keyId: nil, variation: nil, expectedMatch: true},
			"service match":     {serviceId: serviceId, featureId: nil, keyId: nil, variation: nil, expectedMatch: true},
			"nil variation":     {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: true},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: false},
			"different feature": {serviceId: serviceId, featureId: ptr.To(uint(4)), keyId: &keyId, variation: nil, expectedMatch: false},
			"different key":     {serviceId: serviceId, featureId: &featureId, keyId: ptr.To(uint(4)), variation: nil, expectedMatch: false},
		}

		test.RunCases(t, run, testCases)
	})
}

func TestServicePermission(t *testing.T) {
	servicePermission := auth.ServicePermission{
		ServiceID: 1,
		Level:     constants.PermissionAdmin,
	}

	serviceId := uint(1)
	featureId := uint(2)
	keyId := uint(3)

	t.Run("Match", func(t *testing.T) {
		type testCase struct {
			serviceId          uint
			featureId          *uint
			keyId              *uint
			variation          map[uint][]string
			expectedPermission constants.PermissionLevel
		}

		run := func(t *testing.T, tc testCase) {
			permission := servicePermission.Match(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, permission, tc.expectedPermission)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionAdmin},
			"match with nil":    {serviceId: serviceId, featureId: nil, keyId: nil, variation: nil, expectedPermission: constants.PermissionAdmin},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MatchAny", func(t *testing.T) {
		type testCase struct {
			serviceId     uint
			featureId     *uint
			keyId         *uint
			variation     map[uint][]string
			expectedMatch bool
		}

		run := func(t *testing.T, tc testCase) {
			match := servicePermission.MatchAny(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, match, tc.expectedMatch)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedMatch: true},
			"match with nil":    {serviceId: serviceId, featureId: nil, keyId: nil, variation: nil, expectedMatch: true},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: false},
		}

		test.RunCases(t, run, testCases)
	})
}

func TestFeaturePermission(t *testing.T) {
	featurePermission := auth.FeaturePermission{
		ServiceID: 1,
		FeatureID: 2,
		Level:     constants.PermissionAdmin,
	}

	serviceId := uint(1)
	featureId := uint(2)
	keyId := uint(3)

	t.Run("Match", func(t *testing.T) {
		type testCase struct {
			serviceId          uint
			featureId          *uint
			keyId              *uint
			variation          map[uint][]string
			expectedPermission constants.PermissionLevel
		}

		run := func(t *testing.T, tc testCase) {
			permission := featurePermission.Match(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, permission, tc.expectedPermission)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionAdmin},
			"match with nil":    {serviceId: serviceId, featureId: &featureId, keyId: nil, variation: nil, expectedPermission: constants.PermissionAdmin},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"different feature": {serviceId: serviceId, featureId: ptr.To(uint(4)), keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"nil feature":       {serviceId: serviceId, featureId: nil, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MatchAny", func(t *testing.T) {
		type testCase struct {
			serviceId     uint
			featureId     *uint
			keyId         *uint
			variation     map[uint][]string
			expectedMatch bool
		}

		run := func(t *testing.T, tc testCase) {
			match := featurePermission.MatchAny(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, match, tc.expectedMatch)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedMatch: true},
			"feature match":     {serviceId: serviceId, featureId: &featureId, keyId: nil, variation: nil, expectedMatch: true},
			"service match":     {serviceId: serviceId, featureId: nil, keyId: nil, variation: nil, expectedMatch: true},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: false},
			"different feature": {serviceId: serviceId, featureId: ptr.To(uint(4)), keyId: &keyId, variation: nil, expectedMatch: false},
		}

		test.RunCases(t, run, testCases)
	})
}

func TestKeyPermission(t *testing.T) {
	keyPermission := auth.KeyPermission{
		ServiceID: 1,
		FeatureID: 2,
		KeyID:     3,
		Level:     constants.PermissionAdmin,
	}

	serviceId := uint(1)
	featureId := uint(2)
	keyId := uint(3)

	t.Run("Match", func(t *testing.T) {
		type testCase struct {
			serviceId          uint
			featureId          *uint
			keyId              *uint
			variation          map[uint][]string
			expectedPermission constants.PermissionLevel
		}

		run := func(t *testing.T, tc testCase) {
			permission := keyPermission.Match(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, permission, tc.expectedPermission)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedPermission: constants.PermissionAdmin},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"different feature": {serviceId: serviceId, featureId: ptr.To(uint(4)), keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"different key":     {serviceId: serviceId, featureId: &featureId, keyId: ptr.To(uint(4)), variation: nil, expectedPermission: constants.PermissionViewer},
			"nil feature":       {serviceId: serviceId, featureId: nil, keyId: &keyId, variation: nil, expectedPermission: constants.PermissionViewer},
			"nil key":           {serviceId: serviceId, featureId: &featureId, keyId: nil, variation: nil, expectedPermission: constants.PermissionViewer},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MatchAny", func(t *testing.T) {
		type testCase struct {
			serviceId     uint
			featureId     *uint
			keyId         *uint
			variation     map[uint][]string
			expectedMatch bool
		}

		run := func(t *testing.T, tc testCase) {
			match := keyPermission.MatchAny(tc.serviceId, tc.featureId, tc.keyId, tc.variation)
			assert.Equal(t, match, tc.expectedMatch)
		}

		testCases := map[string]testCase{
			"match": {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: map[uint][]string{
				1: {"value1"},
				2: {"value2"},
			}, expectedMatch: true},
			"key match":         {serviceId: serviceId, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: true},
			"feature match":     {serviceId: serviceId, featureId: &featureId, keyId: nil, variation: nil, expectedMatch: true},
			"service match":     {serviceId: serviceId, featureId: nil, keyId: nil, variation: nil, expectedMatch: true},
			"different service": {serviceId: 2, featureId: &featureId, keyId: &keyId, variation: nil, expectedMatch: false},
			"different feature": {serviceId: serviceId, featureId: ptr.To(uint(4)), keyId: &keyId, variation: nil, expectedMatch: false},
			"different key":     {serviceId: serviceId, featureId: &featureId, keyId: ptr.To(uint(4)), variation: nil, expectedMatch: false},
		}

		test.RunCases(t, run, testCases)
	})
}
