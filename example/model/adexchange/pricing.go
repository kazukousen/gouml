package adexchange

// Pricing object
type Pricing struct {
	Price      float64
	PriceModel PriceModel
}

// PriceOfUnit ...
func (p Pricing) PriceOfUnit() float64 {
	switch p.PriceModel {
	case PriceModelCPM:
		return p.Price / 1000
	}
	return p.Price
}

// PriceModel object
type PriceModel string

// enums of pricing model
const (
	PriceModelCPM PriceModel = "cpm"
)
