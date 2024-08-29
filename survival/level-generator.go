package survival

import (
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine"
)

type SurvivalLevelGenerator struct {
	*engine.BaseLevelGenerator
}

func NewSurvLevelGenerator() (*SurvivalLevelGenerator, error) {
	base, err := engine.NewLevelGenerator()
	if err != nil {
		return nil, err
	}
	g := &SurvivalLevelGenerator{BaseLevelGenerator: base}
	return g, nil
}

func (g *SurvivalLevelGenerator) Generate(am engine.AssetManager) (*engine.GeneratorResult, error) {
	res := &engine.GeneratorResult{}
	worldMap, err := g.GenerateWorldMap(am)
	if err != nil {
		return nil, err
	}
	res.WorldMap = worldMap
	return res, nil
}

func (g *SurvivalLevelGenerator) GenerateWorldMap(am engine.AssetManager) (engine.WorldMap, error) {
	worldWidth, worldHeight := g.WorldDimensions()
	screenWidth, screenHeight := g.ScreenDimensions()
	// Base layer
	baseTiles, err := am.Tileset("darkdimension")
	if err != nil {
		return nil, err
	}
	// FIX: Move to constant
	centerMapPosition := cp.Vector{1456, 1456}
	worldMap, err := engine.NewProcHexWorldMap(worldWidth, worldHeight, centerMapPosition)
	if err != nil {
		return nil, err
	}
	if err := worldMap.AddSkyboxLayer(int64(screenWidth), int64(screenHeight), baseTiles); err != nil {
		return nil, err
	}
	if err := worldMap.AddCsvLayer(assets.MapDarkDarkGroundCSV, baseTiles); err != nil {
		return nil, err
	}

	// Setup empty layers
	castleProps, err := am.Tileset("props")
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
