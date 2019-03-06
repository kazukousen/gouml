package analytics

import (
	"time"

	"github.com/kazukousen/gouml/example/model/item"
)

// Orders object
type Orders []Order

func (os Orders) extractWithInRange(dateRange DateRange) Orders {
	dst := make(Orders, 0, len(os))
	for _, order := range os {
		if order.StoredBetween(dateRange) {
			dst = append(dst, order)
		}
	}
	return dst
}

// Order object
type Order struct {
	Pricing    item.Pricing
	StoredTime time.Time
}

// StoredBetween ...
func (o Order) StoredBetween(dateRange DateRange) bool {
	return dateRange.Include(o.StoredTime)
}
