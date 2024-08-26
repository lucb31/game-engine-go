package survival

import (
	"math/rand"

	"github.com/jakecoffman/cp"
)

func entityCircleDistribution(areaCenter cp.Vector, areaRadius float64, entityCount int, spacing float64) []cp.Vector {
	maxTries := entityCount * 10
	entityBBs := []cp.BB{}
	entityPos := []cp.Vector{}
	tries := 0
	for ; len(entityBBs) < entityCount && tries < maxTries; tries++ {
		x := rand.Float64()*areaRadius*2 - areaRadius
		y := rand.Float64()*areaRadius*2 - areaRadius
		currentCenter := areaCenter.Add(cp.Vector{x, y})
		currentBB := cp.NewBBForCircle(currentCenter, spacing)
		intersects := false
		for _, bb := range entityBBs {
			if bb.Intersects(currentBB) {
				intersects = true
				break
			}
		}
		if !intersects {
			entityBBs = append(entityBBs, currentBB)
			entityPos = append(entityPos, currentCenter)
		}
	}
	return entityPos
}
