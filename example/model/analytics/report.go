package analytics

// BeaconEventReport object
type BeaconEventReport struct {
	Events    BeaconEvents
	DateRange DateRange
}

// SummarizeWithInDateRange ...
func (r BeaconEventReport) SummarizeWithInDateRange() (uint, float64) {
	return r.summarize(r.Events.extractWithInRange(r.DateRange))
}

func (r BeaconEventReport) summarize(events BeaconEvents) (count uint, revenue float64) {
	for _, e := range events {
		revenue += e.Pricing.PriceOfUnit()
		count++
	}
	return
}
