package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type Biome int

const (
	Gras  Biome = 32
	Rock  Biome = 42
	Undef Biome = 71
)

// TODO: Should not repeat here
const mapTileSize = 16

type GameWorld struct {
	objects      []GameEntity
	player       GameEntity
	Biome        [][]Biome
	Width        int64
	Height       int64
	FrameCount   int64
	AssetManager *AssetManager
	space        *cp.Space
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	w.drawBiomes(screen)
	// TODO: Currently drawing ALL objects. Fine as long as there is no camera movement
	for _, obj := range w.objects {
		obj.Draw(screen)
	}
	w.player.Draw(screen)
}

func (w *GameWorld) Update() {
	w.FrameCount++
	w.space.Step(1.0 / 60.0)
	//for _, obj := range w.objects {
	//	obj.Update()
	//}
	w.player.Update()
}

func (w *GameWorld) drawBiomes(screen *ebiten.Image) {
	// Drawing WHOLE map. This is ok because there is no camera movement right now
	for row := range w.Height {
		for col := range w.Width {
			// Set tile position
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*mapTileSize), float64(row*mapTileSize))

			// Select correct tile from tileset
			subIm, err := w.AssetManager.GetTile("plains", int(w.Biome[row][col]))
			if err != nil {
				fmt.Println("Unable to draw biome cell", err.Error())
				return
			}
			screen.DrawImage(subIm, &op)
		}
	}
}

func createBiome(width, height int64) ([][]Biome, error) {
	// TODO: Should be coming from file / external source
	mapData := [][]Biome{
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{13, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 15},
	}
	biome := make([][]Biome, height)
	// Copy map data in & fill remaining cells with placeholder tile
	for row := range height {
		biome[row] = make([]Biome, width)
		for col := range width {
			if int64(len(mapData)) > row && int64(len(mapData[row])) > col {
				biome[row][col] = mapData[row][col]
			} else {
				biome[row][col] = Undef
			}
		}
	}
	return biome, nil
}

func createFences(am *AssetManager, space *cp.Space) ([]GameEntity, error) {
	fenceData := [][]int{
		{-1, -1, -1, -1, -1, -1, -1, -1},
		{-1, -1, -1, -1, 1, 14, 14, 3},
		{-1, -1, -1, -1, 4, -1, -1, 4},
		{-1, -1, -1, -1, 4, -1, -1, 4},
		{-1, -1, -1, -1, 9, 14, 14, 10},
	}
	objects := []GameEntity{}
	// Temporary disable fence
	return objects, nil
	for row, rowData := range fenceData {
		for col, tileIdx := range rowData {
			if tileIdx > -1 {
				im, err := am.GetTile("fences", tileIdx)
				if err != nil {
					return nil, err
				}
				body := cp.NewStaticBody()
				body.SetPosition(cp.Vector{X: float64(mapTileSize * col), Y: float64(mapTileSize * row)})
				shape := cp.NewBox(body, 16, 16, 0)
				//shape := cp.NewCircle(body, 8, cp.Vector{})

				shape.SetFriction(1)
				shape.SetElasticity(0)
				space.AddBody(body)
				space.AddShape(shape)
				objects = append(objects, &StaticGameEntity{Shape: shape, Image: im})
			}
		}
	}
	return objects, nil
}

// Static prop in world. Collidable, but no movement, no animation
type StaticGameEntity struct {
	Image *ebiten.Image
	Shape *cp.Shape
}

// Nothing to do since its static
func (p *StaticGameEntity) Update() {}

func (p *StaticGameEntity) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.Shape.Body().Position().X, p.Shape.Body().Position().Y)
	screen.DrawImage(p.Image, &op)
}

func NewWorld(width int64, height int64) (*GameWorld, error) {
	// Initialize assets
	am, err := NewAssetManager()
	if err != nil {
		return nil, err
	}
	// Initialize map
	biome, err := createBiome(width, height)
	if err != nil {
		return nil, err
	}
	// Initialize physics
	space := cp.NewSpace()
	// Initialize bounding box
	offset := 2.5
	walls := []cp.Vector{
		{offset, offset}, {offset, 240},
		{320, offset}, {320, 240},
		{offset, offset}, {320, offset},
		{offset, 240}, {320, 240},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 2))
		shape.SetElasticity(1)
		shape.SetFriction(1)
	}

	// Initialize some fences
	objects, err := createFences(am, space)
	if err != nil {
		return nil, err
	}
	w := GameWorld{Biome: biome, Width: width, Height: height, AssetManager: am, objects: objects, space: space}

	// Initialize player (after world has been initialized to reference it)
	asset, ok := am.CharacterAssets["player"]
	if !ok {
		return nil, fmt.Errorf("Could not find player asset")
	}
	asset.currentFrame = &w.FrameCount
	player, err := NewPlayer(&w, &asset)
	if err != nil {
		return &w, err
	}
	w.player = player
	space.AddBody(player.Shape.Body())
	space.AddShape(player.Shape)

	// Initialize an npc
	np, err := NewNpc(&asset)
	space.AddBody(np.Shape.Body())
	space.AddShape(np.Shape)
	w.objects = append(w.objects, np)
	return &w, nil
}
