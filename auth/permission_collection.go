package auth

import (
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/service"
)

type VariationPropertyValueParentsProvider interface {
	GetParents(propertyId uint, value string) []string
}

type PermissionCollection struct {
	permissions     []Permission
	parentsProvider VariationPropertyValueParentsProvider
}

func NewPermissionCollection(userPermissions []service.UserPermission, parentsProvider VariationPropertyValueParentsProvider) *PermissionCollection {
	permissions := make([]Permission, len(userPermissions))

	for i, permission := range userPermissions {
		permissions[i] = CreatePermission(permission)
	}

	return &PermissionCollection{
		permissions:     permissions,
		parentsProvider: parentsProvider,
	}
}

func CreatePermission(userPermission service.UserPermission) Permission {
	if len(userPermission.Variation) > 0 {
		return &VariationPermission{
			ServiceID:      userPermission.ServiceID,
			FeatureID:      *userPermission.FeatureID,
			KeyID:          *userPermission.KeyID,
			PropertyValues: userPermission.Variation,
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

func (p *PermissionCollection) GetPermissionLevelFor(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint]string) constants.PermissionLevel {
	var variationPropertyValuesWithParents map[uint][]string

	if variationPropertyValues != nil {
		variationPropertyValuesWithParents = make(map[uint][]string, len(variationPropertyValues))

		for propertyId, propertyValue := range variationPropertyValues {
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
