package item

// Pricing object
type Pricing struct {
	Price    float64
	Currency PriceCurrency
}

// PriceOfUnit ...
func (p Pricing) PriceOfUnit() float64 {
	switch p.Currency {
	case PriceCurrencyBTC:
		return p.Price / 100000
	}
	return p.Price
}

// PriceCurrency object
type PriceCurrency string

// enums of price currency
const (
	PriceCurrencyJPY PriceCurrency = "JPY"
	PriceCurrencyBTC PriceCurrency = "BTC"
)
