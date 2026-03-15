package domain

// Footing represents the physical dimensions of a concrete footing.
// Keeping it simple as a rectangular cuboid for the initial MVP.
type Footing struct {
	Name   string
	Length float64 // in meters
	Width  float64 // in meters
	Depth  float64 // in meters
}

// Volume calculates the "wet volume" of the concrete required.
func (f Footing) Volume() float64 {
	return f.Length * f.Width * f.Depth
}
