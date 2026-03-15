package usecase

import (
	"github.com/Govind-619/xtmator/domain"
)

// EstimateResult holds the final calculated quantities and costs for the UI to display.
// This formats the raw domain data into something useful for the user.
type EstimateResult struct {
	FootingVolume float64
	CementBags    float64
	CementCost    float64
	SandVolume    float64
	SandCost      float64
	AggrVolume    float64
	AggrCost      float64
	TotalCost     float64
}

// EstimateFootingCost orchestrates the domain logic to generate a full estimate.
// Notice how it relies purely on the domain structs we just built!
func EstimateFootingCost(
	f domain.Footing,
	mix domain.ConcreteMix,
	cement domain.Material,
	sand domain.Material,
	aggr domain.Material,
) EstimateResult {

	// 1. Calculate the wet volume using the Footing domain logic
	vol := f.Volume()

	// 2. Calculate the required material quantities using the Mix Design domain logic
	quants := mix.CalculateMaterials(vol)

	// 3. Calculate the individual costs using the Material domain logic
	cementCost := quants.CementBags * cement.Price
	sandCost := quants.SandVolume * sand.Price
	aggrCost := quants.AggrVolume * aggr.Price

	// 4. Compile the final result
	return EstimateResult{
		FootingVolume: vol,
		CementBags:    quants.CementBags,
		CementCost:    cementCost,
		SandVolume:    quants.SandVolume,
		SandCost:      sandCost,
		AggrVolume:    quants.AggrVolume,
		AggrCost:      aggrCost,
		TotalCost:     cementCost + sandCost + aggrCost,
	}
}
