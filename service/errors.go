package service

import "errors"

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrInvalidPassword = errors.New("invalid password")
)
