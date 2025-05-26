package auth

import (
	"github.com/necroskillz/config-service/constants"
)

type VariationPropertyValueParentsProvider interface {
	GetParents(propertyId uint, value string) ([]string, error)
}

type PermissionCollection struct {
	permissions     []Permission
	parentsProvider VariationPropertyValueParentsProvider
}

func NewPermissionCollection(parentsProvider VariationPropertyValueParentsProvider) *PermissionCollection {
	permissions := []Permission{}

	return &PermissionCollection{
		permissions:     permissions,
		parentsProvider: parentsProvider,
	}
}

func (p *PermissionCollection) AddPermission(serviceId uint, featureId *uint, keyId *uint, variation map[uint]string, permissionLevel constants.PermissionLevel) Permission {
	var permission Permission

	if len(variation) > 0 {
		permission = &VariationPermission{
			ServiceID:      serviceId,
			FeatureID:      *featureId,
			KeyID:          *keyId,
			PropertyValues: variation,
			Level:          permissionLevel,
		}
	} else if keyId != nil {
		permission = &KeyPermission{
			ServiceID: serviceId,
			FeatureID: *featureId,
			KeyID:     *keyId,
			Level:     permissionLevel,
		}
	} else if featureId != nil {
		permission = &FeaturePermission{
			ServiceID: serviceId,
			FeatureID: *featureId,
			Level:     permissionLevel,
		}
	} else {
		permission = &ServicePermission{
			ServiceID: serviceId,
			Level:     permissionLevel,
		}
	}

	p.permissions = append(p.permissions, permission)

	return permission
}

func (p *PermissionCollection) GetPermissionLevelFor(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint]string) constants.PermissionLevel {
	var variationPropertyValuesWithParents map[uint][]string

	if variationPropertyValues != nil {
		variationPropertyValuesWithParents = make(map[uint][]string, len(variationPropertyValues))

		for propertyId, propertyValue := range variationPropertyValues {
			parents, err := p.parentsProvider.GetParents(propertyId, propertyValue)
			if err != nil {
				// TODO: log error
				return constants.PermissionViewer
			}

			variationPropertyValuesWithParents[propertyId] = append(parents, propertyValue)
		}
	}

	maxPermissionLevel := constants.PermissionViewer

	for _, permission := range p.permissions {
		permissionLevel := permission.Match(serviceId, featureId, keyId, variationPropertyValuesWithParents)

		maxPermissionLevel = max(maxPermissionLevel, permissionLevel)
	}

	return maxPermissionLevel
}

func (p *PermissionCollection) HasPermissionForNestedEntity(serviceId uint, featureId *uint, keyId *uint) bool {
	for _, permission := range p.permissions {
		if permission.MatchAny(serviceId, featureId, keyId, nil) {
			return true
		}
	}

	return false
}
