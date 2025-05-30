package auth

import "github.com/necroskillz/config-service/constants"

type UserBuilder struct {
	user *User
}

func NewUserBuilder(parentsProvider VariationPropertyValueParentsProvider) *UserBuilder {
	return &UserBuilder{
		user: &User{
			permissionCollection: NewPermissionCollection(parentsProvider),
		},
	}
}

func (u *UserBuilder) User() *User {
	return u.user
}

func (u *UserBuilder) WithBasicInfo(id uint, username string, isGlobalAdmin bool) *UserBuilder {
	u.user.ID = id
	u.user.Username = username
	u.user.IsGlobalAdmin = isGlobalAdmin
	u.user.IsAuthenticated = true
	return u
}

func (u *UserBuilder) WithChangesetID(changesetID uint) *UserBuilder {
	u.user.ChangesetID = changesetID
	return u
}

func (u *UserBuilder) WithPermission(serviceId uint, featureId *uint, keyId *uint, variation map[uint]string, permission constants.PermissionLevel) *UserBuilder {
	u.user.permissionCollection.AddPermission(serviceId, featureId, keyId, variation, permission)
	return u
}
