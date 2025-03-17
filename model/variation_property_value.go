package model

type VariationPropertyValue struct {
	ID                  uint
	VariationProperty   VariationProperty
	VariationPropertyID uint
	Value               string
	Parent              *VariationPropertyValue
	ParentID            *uint
}
