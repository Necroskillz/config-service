package auth

import (
	"github.com/necroskillz/config-service/constants"
)

type User struct {
	ID                   uint
	Username             string
	IsAuthenticated      bool
	IsGlobalAdmin        bool
	permissionCollection *PermissionCollection
	ChangesetID          uint
}

func AnonymousUser() *User {
	return &User{
		ID:              0,
		Username:        "Anonymous",
		IsAuthenticated: false,
		IsGlobalAdmin:   false,
	}
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

func (u *User) HasPermissionForNestedEntity(serviceId uint, featureId uint, keyId uint) bool {
	if !u.IsAuthenticated {
		return false
	}

	if u.IsGlobalAdmin {
		return true
	}

	return u.permissionCollection.HasPermissionForNestedEntity(serviceId, &featureId, &keyId)
}
