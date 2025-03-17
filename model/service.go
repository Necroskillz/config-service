package model

import (
	"time"
)

type Service struct {
	ID            uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string `gorm:"uniqueIndex"`
	Description   string
	Archived      bool
	ServiceType   ServiceType
	ServiceTypeID uint
}
