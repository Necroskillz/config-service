package model

import (
	"time"
)

type FeatureVersion struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	ValidFrom *time.Time `gorm:"index"`
	ValidTo   *time.Time `gorm:"index"`
	Version   int
	Feature   Feature
	FeatureID uint
	Archived  bool
}
