package model

import (
	"time"
)

type FeatureVersionServiceVersion struct {
	ID               uint
	CreatedAt        time.Time
	ValidFrom        *time.Time `gorm:"index"`
	ValidTo          *time.Time `gorm:"index"`
	FeatureVersion   FeatureVersion
	FeatureVersionID uint
	ServiceVersion   ServiceVersion
	ServiceVersionID uint
}
