package analytics

import (
	"time"

	"github.com/kazukousen/gouml/example/model/adexchange"
)

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
