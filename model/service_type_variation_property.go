package model

import "time"

type ServiceTypeVariationProperty struct {
	ID                  uint
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Priority            int
	ServiceTypeID       uint
	ServiceType         ServiceType
	VariationPropertyID uint
	VariationProperty   VariationProperty
}
