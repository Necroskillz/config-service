package model

import (
	"time"
)

type Feature struct {
	ID          uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string `gorm:"uniqueIndex"`
	Description string
	Archived    bool
	Service     Service
	ServiceID   uint
}
