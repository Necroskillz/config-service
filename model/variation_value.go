package model

import (
	"time"
)

type VariationValue struct {
	ID                      uint
	ValidFrom               *time.Time `gorm:"index"`
	ValidTo                 *time.Time `gorm:"index"`
	Key                     Key
	KeyID                   uint
	Data                    *string
	VariationPropertyValues []VariationPropertyValue `gorm:"many2many:variation_value_variation_property_values;constraint:OnDelete:CASCADE;"`
}
