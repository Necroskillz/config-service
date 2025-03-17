package auth

import (
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
)

type User struct {
	ID                   uint
	Username             string
	IsAuthenticated      bool
	IsGlobalAdmin        bool
	permissionCollection *PermissionCollection
}

func AnonymousUser() *User {
	return &User{
		ID:              uint(0),
		Username:        "Anonymous",
		IsAuthenticated: false,
		IsGlobalAdmin:   false,
	}
}

func NewUser(model *model.User, parentsProvider VariationPropertyValueParentsProvider) *User {
	user := &User{
		ID:                   model.ID,
		Username:             model.Name,
		IsAuthenticated:      true,
		IsGlobalAdmin:        model.GlobalAdministrator,
		permissionCollection: NewPermissionCollection(model.Permissions, parentsProvider),
	}

	return user
}

func (u *User) getPermission(serviceId uint, featureId *uint, keyId *uint, variationPropertyValues map[uint]string) constants.PermissionLevel {
	if !u.IsAuthenticated {
		return constants.PermissionViewer
	}

	if u.IsGlobalAdmin {
		return constants.PermissionAdmin
	}

	return u.permissionCollection.GetPermissionLevelFor(serviceId, featureId, keyId, variationPropertyValues)
}

func (u *User) GetPermissionForService(serviceId uint) constants.PermissionLevel {
	return u.getPermission(serviceId, nil, nil, nil)
}

func (u *User) GetPermissionForFeature(serviceId uint, featureId uint) constants.PermissionLevel {
	return u.getPermission(serviceId, &featureId, nil, nil)
}

func (u *User) GetPermissionForKey(serviceId uint, featureId uint, keyId uint) constants.PermissionLevel {
	return u.getPermission(serviceId, &featureId, &keyId, nil)
}

func (u *User) GetPermissionForValue(serviceId uint, featureId uint, keyId uint, variationPropertyValues map[uint]string) constants.PermissionLevel {
	return u.getPermission(serviceId, &featureId, &keyId, variationPropertyValues)
}
