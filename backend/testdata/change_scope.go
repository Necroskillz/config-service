package testdata

type ChangeAmount struct {
	base    int
	variant int
}

func NewChangeAmount(min int, max int) ChangeAmount {
	return ChangeAmount{
		base:    min,
		variant: max - min,
	}
}

func (c *ChangeAmount) Next(rng *Rng) int {
	if c.variant == 0 {
		return c.base
	}

	return c.base + rng.Intn(c.variant+1)
}

type ChangeScope struct {
	Weight               int
	CreateValue          ChangeAmount
	UpdateValue          ChangeAmount
	DeleteValue          ChangeAmount
	CreateKey            ChangeAmount
	DeleteKey            ChangeAmount
	CreateFeature        ChangeAmount
	LinkFeature          ChangeAmount
	UnlinkFeature        ChangeAmount
	CreateFeatureVersion ChangeAmount
	CreateServiceVersion ChangeAmount
}
