package model

type VariationProperty struct {
	ID           uint
	Name         string
	Priority     int
	Values       []VariationPropertyValue
	ServiceTypes []ServiceTypeVariationProperty
}
