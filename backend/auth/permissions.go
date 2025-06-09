package auth

import (
	"slices"

	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/util/ptr"
)

type Permission interface {
	Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel
	MatchAny(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) bool
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

func (p *ServicePermission) MatchAny(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) bool {
	return p.ServiceID == serviceId
}

type FeaturePermission struct {
	ServiceID uint
	FeatureID uint
	Level     constants.PermissionLevel
}

func (p *FeaturePermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId && p.FeatureID == ptr.From(featureId) {
		return p.Level
	}

	return constants.PermissionViewer
}

func (p *FeaturePermission) MatchAny(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) bool {
	match := p.ServiceID == serviceId

	if match && featureId != nil {
		match = match && p.FeatureID == ptr.From(featureId)
	}

	return match
}

type KeyPermission struct {
	ServiceID uint
	FeatureID uint
	KeyID     uint
	Level     constants.PermissionLevel
}

func (p *KeyPermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId && p.FeatureID == ptr.From(featureId) && p.KeyID == ptr.From(keyId) {
		return p.Level
	}

	return constants.PermissionViewer
}

func (p *KeyPermission) MatchAny(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) bool {
	match := p.ServiceID == serviceId

	if match && featureId != nil {
		match = match && p.FeatureID == ptr.From(featureId)
	}

	if match && keyId != nil {
		match = match && p.KeyID == ptr.From(keyId)
	}

	return match
}

type VariationPermission struct {
	ServiceID      uint
	FeatureID      uint
	KeyID          uint
	PropertyValues map[uint]string
	Level          constants.PermissionLevel
}

func (p *VariationPermission) Match(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) constants.PermissionLevel {
	if p.ServiceID == serviceId && p.FeatureID == ptr.From(featureId) && p.KeyID == ptr.From(keyId) {
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

func (p *VariationPermission) MatchAny(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint][]string) bool {
	match := p.ServiceID == serviceId

	if match && featureId != nil {
		match = match && p.FeatureID == ptr.From(featureId)
	}

	if match && keyId != nil {
		match = match && p.KeyID == ptr.From(keyId)
	}

	return match
}
