package model

type VariationProperty struct {
	ID           uint
	Name         string
	DisplayName  string
	Priority     int
	Values       []VariationPropertyValue
	ServiceTypes []ServiceTypeVariationProperty
}
