package model

import "github.com/necroskillz/config-service/constants"

type UserPermission struct {
	ID                      uint
	UserID                  uint
	User                    User
	ServiceID               uint
	FeatureID               *uint
	KeyID                   *uint
	VariationPropertyValues []VariationPropertyValue `gorm:"many2many:user_permission_variation_property_values;"`
	Permission              constants.PermissionLevel
}
