package model

import (
	"time"
)

type ServiceVersion struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	ValidFrom *time.Time `gorm:"index"`
	ValidTo   *time.Time `gorm:"index"`
	Service   Service
	ServiceID uint
	Version   int
	Features  []FeatureVersionServiceVersion
	Published bool
	Archived  bool
}
