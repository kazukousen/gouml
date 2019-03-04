package analytics

import "time"

// DateRange object
type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// Include ...
func (dr DateRange) Include(t time.Time) bool {
	if dr.StartDate.After(t) && dr.EndDate.Before(t) {
		return true
	}

	return false
}
