package survival

import (
	"math/rand"
	"slices"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine"
)

type SurvivalLevelGenerator struct {
	*engine.BaseLevelGenerator
	am engine.AssetManager
}

var centerMapPosition = cp.Vector{1456, 1456}

func NewSurvLevelGenerator() (*SurvivalLevelGenerator, error) {
	base, err := engine.NewLevelGenerator()
	if err != nil {
		return nil, err
	}
	g := &SurvivalLevelGenerator{BaseLevelGenerator: base}
	return g, nil
}

func (g *SurvivalLevelGenerator) Generate(am engine.AssetManager) (*engine.GeneratorResult, error) {
	g.am = am
	res := &engine.GeneratorResult{}
	// Generate map
	worldMap, err := g.GenerateWorldMap()
	if err != nil {
		return nil, err
	}
	res.WorldMap = worldMap

	// Generate forest
	treeObjects, err := g.GenerateForest()
	if err != nil {
		return nil, err
	}
	res.Objects = treeObjects

	return res, nil
}

var availableTrees = []string{"tree_a", "tree_b", "tree_small"}

// Spawn a bunch of random trees around center position
func (g *SurvivalLevelGenerator) GenerateTreesAroundCenter(center cp.Vector) ([]engine.GameEntity, error) {
	treeCount := 800
	treeRadius := 32.0
	treePositions := entityDonutDistribution(center, 500, 1200, treeCount, treeRadius)
	res := []engine.GameEntity{}
	for _, pos := range treePositions {
		treeIdx := rand.Intn(len(availableTrees))
		treeType := availableTrees[treeIdx]
		asset, err := g.am.CharacterAsset(treeType)
		if err != nil {
			return []engine.GameEntity{}, err
		}
		var tree *engine.TreeEntity
		if treeType == "tree_small" {
			tree, err = engine.NewBush(asset)
		} else {
			tree, err = engine.NewTree(asset)
		}
		if err != nil {
			return []engine.GameEntity{}, err
		}
		tree.SetPosition(pos)
		res = append(res, tree)
	}
	return res, nil
}

func (g *SurvivalLevelGenerator) GenerateForest() ([]engine.GameEntity, error) {
	res := []engine.GameEntity{}
	// Generate trees around starting segment
	aroundStartingHex, err := g.GenerateTreesAroundCenter(centerMapPosition)
	if err != nil {
		return nil, err
	}
	// Sort objects by vertical position to create 2.5d effect
	// Could possibly move this to render function instead
	slices.SortFunc(aroundStartingHex, func(a, b engine.GameEntity) int {
		return int(a.Shape().Body().Position().Y) - int(b.Shape().Body().Position().Y)
	})
	res = append(res, aroundStartingHex...)
	return res, nil
}

func (g *SurvivalLevelGenerator) GenerateWorldMap() (engine.WorldMap, error) {
	worldWidth, worldHeight := g.WorldDimensions()
	screenWidth, screenHeight := g.ScreenDimensions()

	// Base layer
	baseTiles, err := g.am.Tileset("darkdimension")
	if err != nil {
		return nil, err
	}
	worldMap, err := engine.NewProcHexWorldMap(worldWidth, worldHeight, centerMapPosition)
	if err != nil {
		return nil, err
	}
	if err := worldMap.AddSkyboxLayer(int64(screenWidth), int64(screenHeight), baseTiles); err != nil {
		return nil, err
	}
	// if err := worldMap.AddCsvLayer(assets.MapDarkDarkGroundCSV, baseTiles); err != nil {
	// 	return nil, err
	// }

	// Setup empty layers
	castleProps, err := g.am.Tileset("props")
	if err != nil {
		return nil, err
	}
	if err := worldMap.InitHexBaseLayers(castleProps); err != nil {
		return nil, err
	}

	// Add to segment pool
	if err := worldMap.AddHexSegment(assets.Hex128112CSV); err != nil {
		return nil, err
	}
	if err := worldMap.AddHexSegment(assets.Hex128112PoolBaseCSV); err != nil {
		return nil, err
	}

	// Generate map
	if err := worldMap.Generate(); err != nil {
		return nil, err
	}

	// Temporarily disable castle props & collision layers
	return worldMap, nil
	// Inner walls layer
	// if err := worldMap.AddCollisionLayer(assets.MapDarkLogicWallsCSV); err != nil {
	// 	return nil, err
	// }

	// Castle front layer
	if err := worldMap.AddCsvLayer(assets.MapDarkDarkCastleWallsCSV, baseTiles); err != nil {
		return nil, err
	}

	if err := worldMap.AddCsvLayer(assets.MapDarkPropsPropsCSV, castleProps); err != nil {
		return nil, err
	}

	// Outside layer
	if err := worldMap.AddCsvLayer(assets.MapDarkDarkOutsidePropsCSV, baseTiles); err != nil {
		return nil, err
	}

	return worldMap, nil
}
