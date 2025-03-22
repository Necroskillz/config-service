package auth

import (
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
)

type VariationPropertyValueParentsProvider interface {
	GetParents(propertyId uint, value string) []string
	GetPropertyId(property string) (uint, error)
}

type PermissionCollection struct {
	permissions     []Permission
	parentsProvider VariationPropertyValueParentsProvider
}

func NewPermissionCollection(userPermissions []model.UserPermission, parentsProvider VariationPropertyValueParentsProvider) *PermissionCollection {
	permissions := make([]Permission, len(userPermissions))

	for i, permission := range userPermissions {
		permissions[i] = CreatePermission(permission)
	}

	return &PermissionCollection{
		permissions:     permissions,
		parentsProvider: parentsProvider,
	}
}

func CreatePermission(userPermission model.UserPermission) Permission {
	if len(userPermission.VariationPropertyValues) > 0 {
		propertyValues := make(map[uint]string)

		for _, propertyValue := range userPermission.VariationPropertyValues {
			propertyValues[propertyValue.VariationPropertyID] = propertyValue.Value
		}

		return &VariationPermission{
			ServiceID:      userPermission.ServiceID,
			FeatureID:      *userPermission.FeatureID,
			KeyID:          *userPermission.KeyID,
			PropertyValues: propertyValues,
			Level:          userPermission.Permission,
		}
	} else if userPermission.KeyID != nil {
		return &KeyPermission{
			ServiceID: userPermission.ServiceID,
			FeatureID: *userPermission.FeatureID,
			KeyID:     *userPermission.KeyID,
			Level:     userPermission.Permission,
		}
	} else if userPermission.FeatureID != nil {
		return &FeaturePermission{
			ServiceID: userPermission.ServiceID,
			FeatureID: *userPermission.FeatureID,
			Level:     userPermission.Permission,
		}
	} else {
		return &ServicePermission{
			ServiceID: userPermission.ServiceID,
			Level:     userPermission.Permission,
		}
	}
}

func (p *PermissionCollection) GetPermissionLevelFor(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[string]string) constants.PermissionLevel {
	var variationPropertyValuesWithParents map[uint][]string

	if variationPropertyValues != nil {
		variationPropertyValuesWithParents = make(map[uint][]string, len(variationPropertyValues))

		for property, propertyValue := range variationPropertyValues {
			propertyId, err := p.parentsProvider.GetPropertyId(property)

			if err != nil {
				panic(err)
			}

			parents := p.parentsProvider.GetParents(propertyId, propertyValue)
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
