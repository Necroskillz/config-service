package auth_test

import (
	"testing"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"gotest.tools/v3/assert"
)

func TestVariationPermission(t *testing.T) {
	variationPermission := auth.VariationPermission{
		ServiceID: 1,
		FeatureID: 1,
		KeyID:     1,
		Level:     constants.PermissionAdmin,
		PropertyValues: map[uint]string{
			1: "value1",
			2: "value2",
		},
	}

	serviceId := uint(1)
	featureId := uint(1)
	keyId := uint(1)

	assert.Equal(t, constants.PermissionAdmin, variationPermission.Match(serviceId, &featureId, &keyId, map[uint][]string{
		1: {"value", "value1"},
		2: {"value2"},
	}))

	assert.Equal(t, constants.PermissionViewer, variationPermission.Match(serviceId, &featureId, &keyId, map[uint][]string{
		1: {"value1"},
		3: {"value2"},
	}))

	assert.Equal(t, constants.PermissionAdmin, variationPermission.Match(serviceId, &featureId, &keyId, map[uint][]string{
		1: {"value1"},
		2: {"value2"},
		3: {"value3"},
	}))
}
