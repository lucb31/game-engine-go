package engine

type GeneratorResult struct {
	WorldMap WorldMap
	Objects  []GameEntity
}

type WorldGenerator interface {
	WorldDimensions() (int64, int64)
	// Internal interface every generator has to implement
	Generate(AssetManager) (*GeneratorResult, error)
}

type BaseLevelGenerator struct {
	worldWidth, worldHeight   int64
	screenWidth, screenHeight int
}

func NewLevelGenerator() (*BaseLevelGenerator, error) {
	return &BaseLevelGenerator{}, nil
}

func (g *BaseLevelGenerator) WorldDimensions() (int64, int64) { return g.worldWidth, g.worldHeight }
func (g *BaseLevelGenerator) ScreenDimensions() (int, int)    { return g.screenWidth, g.screenHeight }
func (g *BaseLevelGenerator) SetWorldDimensions(w, h int64)   { g.worldWidth, g.worldHeight = w, h }
func (g *BaseLevelGenerator) SetScreenDimension(w, h int)     { g.screenWidth, g.screenHeight = w, h }
