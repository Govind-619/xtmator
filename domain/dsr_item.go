package domain

// DSRItem represents one item in the District Schedule of Rates catalogue.
type DSRItem struct {
	ID          int64
	Category    string  // e.g. "PCC", "RCC"
	Code        string  // e.g. "4.1.2"
	Description string
	Unit        string  // e.g. "CUM", "SQM"
	Rate        float64 // ₹ per unit (total analysed rate)
}
