package model

import (
	"time"
)

type Key struct {
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ValidFrom        *time.Time `gorm:"index"`
	ValidTo          *time.Time `gorm:"index"`
	Name             string
	Description      *string
	ValueType        ValueType
	ValueTypeID      uint
	FeatureVersion   FeatureVersion
	FeatureVersionID uint
	Values           []VariationValue
}
