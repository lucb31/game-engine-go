package survival

import (
	"log"
	"math"
	"math/rand"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

func GenerateForest(center cp.Vector, am engine.AssetManager, em engine.GameEntityManager) error {
	// Load tree asset(s)
	availableTrees := []string{"tree_a", "tree_b", "tree_small"}

	// Spawn a bunch of random trees in proximity of the castle at 1450, 1000
	treeCount := 2000
	treeRadius := 32.0
	treePositions := entityDonutDistribution(center, 500, 1500, treeCount, treeRadius)
	for _, pos := range treePositions {
		treeIdx := rand.Intn(len(availableTrees))
		treeType := availableTrees[treeIdx]
		asset, err := am.CharacterAsset(treeType)
		if err != nil {
			return err
		}
		var tree *engine.TreeEntity
		if treeType == "tree_small" {
			tree, err = engine.NewBush(em, asset)
		} else {
			tree, err = engine.NewTree(em, asset)
		}
		if err != nil {
			return err
		}
		tree.SetPosition(pos)
		em.AddEntity(tree)
	}
	return nil
}

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
