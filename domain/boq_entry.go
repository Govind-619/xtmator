package domain

// BOQEntry is a single line item in a project's Bill of Quantities.
type BOQEntry struct {
	ID          int64
	ProjectID   int64
	ItemNo      int
	DSRItemID   *int64  // nil for custom/manual entries
	DSRItemCode string
	Description string
	Category    string
	Length      float64
	Breadth     float64
	Height      float64
	Quantity    float64 // CUM — computed L×B×H or manually entered
	Unit        string
	Rate        float64 // ₹ per unit — from DSR or manual override
	Amount      float64 // Quantity × Rate
}
