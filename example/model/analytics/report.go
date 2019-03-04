package analytics

// BeaconEventReport object
type BeaconEventReport struct {
	Events    []BeaconEvent
	DateRange DateRange
}

// SummarizeWithInDateRange ...
func (r BeaconEventReport) SummarizeWithInDateRange() (uint, float64) {
	return r.summarize(r.getEventsWithInRange())
}

func (r BeaconEventReport) summarize(events []BeaconEvent) (count uint, revenue float64) {
	for _, e := range events {
		revenue += e.Pricing.PriceOfUnit()
		count++
	}
	return
}

func (r BeaconEventReport) getEventsWithInRange() []BeaconEvent {
	dst := make([]BeaconEvent, 0, len(r.Events))
	for _, event := range r.Events {
		if event.FiredBetween(r.DateRange) {
			dst = append(dst, event)
		}
	}
	return dst
}
