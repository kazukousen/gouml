package analytics

import (
	"time"

	"github.com/kazukousen/gouml/example/model/adexchange"
)

// BeaconEvents object
type BeaconEvents []BeaconEvent

func (r BeaconEvents) extractWithInRange(dateRange DateRange) BeaconEvents {
	dst := make(BeaconEvents, 0, len(r))
	for _, event := range r {
		if event.FiredBetween(dateRange) {
			dst = append(dst, event)
		}
	}
	return dst
}

// BeaconEvent object
type BeaconEvent struct {
	Code      BeaconEventCode
	Pricing   adexchange.Pricing
	FiredTime time.Time
}

// FiredBetween ...
func (e BeaconEvent) FiredBetween(dateRange DateRange) bool {
	return dateRange.Include(e.FiredTime)
}

// BeaconEventCode object
type BeaconEventCode string
