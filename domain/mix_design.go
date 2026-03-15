package domain

// Standard Civil Engineering Constants
const (
	DryVolumeFactor = 1.54   // Dry volume of concrete is ~54% more than wet volume
	CementDensity   = 1440.0 // kg per cubic meter
	CementBagWeight = 50.0   // kg per bag
)

// ConcreteMix represents a standard mix ratio (e.g., M20 is 1 : 1.5 : 3)
type ConcreteMix struct {
	Grade      string  // e.g., "M20"
	CementPart float64 // e.g., 1.0
	SandPart   float64 // e.g., 1.5
	AggrPart   float64 // e.g., 3.0
}

// ConcreteQuantities holds the final calculated materials needed
type ConcreteQuantities struct {
	CementBags float64 // in Bags (50kg)
	SandVolume float64 // in Cubic Meters
	AggrVolume float64 // in Cubic Meters
}

// CalculateMaterials takes the wet volume (e.g., from a Footing) and calculates exact materials.
func (m ConcreteMix) CalculateMaterials(wetVolume float64) ConcreteQuantities {
	dryVolume := wetVolume * DryVolumeFactor
	totalParts := m.CementPart + m.SandPart + m.AggrPart

	// 1. Calculate individual volumes based on their ratio
	cementVolume := (m.CementPart / totalParts) * dryVolume
	sandVolume := (m.SandPart / totalParts) * dryVolume
	aggrVolume := (m.AggrPart / totalParts) * dryVolume

	// 2. Convert Cement Volume (m3) to Bags
	cementWeightKg := cementVolume * CementDensity
	cementBags := cementWeightKg / CementBagWeight

	return ConcreteQuantities{
		CementBags: cementBags,
		SandVolume: sandVolume,
		AggrVolume: aggrVolume,
	}
}
