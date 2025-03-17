package auth

import (
	"slices"

	"github.com/necroskillz/config-service/constants"
)

type Permission interface {
	Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel
}

type ServicePermission struct {
	ServiceID uint
	Level     constants.PermissionLevel
}

func (p *ServicePermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId {
		return p.Level
	}

	return constants.PermissionViewer
}

type FeaturePermission struct {
	ServiceID uint
	FeatureID uint
	Level     constants.PermissionLevel
}

func (p *FeaturePermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId && p.FeatureID == *featureId {
		return p.Level
	}

	return constants.PermissionViewer
}

type KeyPermission struct {
	ServiceID uint
	FeatureID uint
	KeyID     uint
	Level     constants.PermissionLevel
}

func (p *KeyPermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId && p.FeatureID == *featureId && p.KeyID == *keyId {
		return p.Level
	}

	return constants.PermissionViewer
}

type VariationPermission struct {
	ServiceID      uint
	FeatureID      uint
	KeyID          uint
	PropertyValues map[uint]string
	Level          constants.PermissionLevel
}

func (p *VariationPermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId && p.FeatureID == *featureId && p.KeyID == *keyId {
		for propertyId, permissionPropertyValue := range p.PropertyValues {
			variationPropertyValue, ok := variationPropertyValues[propertyId]

			if !ok || !slices.Contains(variationPropertyValue, permissionPropertyValue) {
				return constants.PermissionViewer
			}
		}

		return p.Level
	}

	return constants.PermissionViewer
}
