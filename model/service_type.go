package model

import "time"

type ServiceType struct {
	ID                  uint
	CreatedAt           time.Time
	Name                string
	VariationProperties []ServiceTypeVariationProperty
}
