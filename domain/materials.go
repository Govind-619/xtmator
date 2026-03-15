package domain

// Material represents a raw building material.
// We keep this simple. The database will eventually populate these.
type Material struct {
	ID    string
	Name  string  // e.g., "Cement", "River Sand", "20mm Aggregate"
	Unit  string  // e.g., "Bags", "m3", "kg"
	Price float64 // Base price per unit in your local currency
}
