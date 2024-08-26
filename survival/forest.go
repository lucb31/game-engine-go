package survival

import (
	"log"
	"math"
	"math/rand"

	"github.com/jakecoffman/cp"
)

func entityDonutDistribution(center cp.Vector, innerRadius, outerRadius float64, count int, spacing float64) []cp.Vector {
	if innerRadius > outerRadius {
		log.Println("Inner radius < Outer radius. Probably not intended!")
	}
	maxTries := count * 10
	entityBBs := []cp.BB{}
	entityPos := []cp.Vector{}
	tries := 0
	for ; len(entityBBs) < count && tries < maxTries; tries++ {
		// 		x := rand.Float64()*areaRadius*2 - areaRadius
		// 		y := rand.Float64()*areaRadius*2 - areaRadius
		radius := innerRadius + rand.Float64()*(outerRadius-innerRadius)
		x := math.Sin(float64(tries)) * radius
		y := math.Cos(float64(tries)) * radius
		currentCenter := center.Add(cp.Vector{x, y})
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
	if tries == maxTries {
		log.Println("Max tries reached! Managed to spawn ", len(entityPos))
	}
	return entityPos
}

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
