package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name                string `gorm:"uniqueIndex"`
	Password            string
	GlobalAdministrator bool
	Permissions         []UserPermission
}
