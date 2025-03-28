package service

import (
	"errors"

	"github.com/necroskillz/config-service/constants"
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrInvalidPassword = errors.New("invalid password")
)

type PermissionChecker interface {
	GetPermissionForService(serviceId uint) constants.PermissionLevel
	GetPermissionForFeature(serviceId uint, featureId uint) constants.PermissionLevel
	GetPermissionForKey(serviceId uint, featureId uint, keyId uint) constants.PermissionLevel
	GetPermissionForValue(serviceId uint, featureId uint, keyId uint, variation map[uint]string) constants.PermissionLevel
}
