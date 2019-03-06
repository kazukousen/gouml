package analytics

// OrderReport object
type OrderReport struct {
	Orders    Orders
	DateRange DateRange
}

// SummarizeWithInDateRange ...
func (r OrderReport) SummarizeWithInDateRange() (uint, float64) {
	return r.summarize(r.Orders.extractWithInRange(r.DateRange))
}

func (r OrderReport) summarize(orders Orders) (count uint, revenue float64) {
	for _, o := range orders {
		revenue += o.Pricing.PriceOfUnit()
		count++
	}
	return
}
