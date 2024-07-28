package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
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
	for _, obj := range w.objects {
		obj.Update()
	}
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
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{13, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 15},
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

func createFences(am *AssetManager) ([]GameEntity, error) {
	fenceData := [][]int{
		{-1, -1, -1, -1, -1, -1, -1, -1},
		{-1, -1, -1, -1, 1, 14, 14, 3},
		{-1, -1, -1, -1, 4, -1, -1, 4},
		{-1, -1, -1, -1, 4, -1, -1, 4},
		{-1, -1, -1, -1, 9, 14, 14, 10},
	}
	objects := []GameEntity{}
	for row, rowData := range fenceData {
		for col, tileIdx := range rowData {
			if tileIdx > -1 {
				im, err := am.GetTile("fences", tileIdx)
				if err != nil {
					return nil, err
				}
				objects = append(objects, &StaticGameEntity{posX: float64(mapTileSize * col), posY: float64(mapTileSize * row), Image: im})
			}
		}
	}
	return objects, nil
}

// Static prop in world. Collidable, but no movement, no animation
type StaticGameEntity struct {
	Image *ebiten.Image
	posX  float64
	posY  float64
}

// Nothing to do since its static
func (p *StaticGameEntity) Update() {}

func (p *StaticGameEntity) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.posX, p.posY)
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
	// Initialize some fences
	objects, err := createFences(am)
	if err != nil {
		return nil, err
	}
	w := GameWorld{Biome: biome, Width: width, Height: height, AssetManager: am, objects: objects}

	// Initialize player (after world has been initialized to reference it)
	player, err := NewPlayer(&w)
	if err != nil {
		return &w, nil
	}
	w.player = player
	return &w, nil
}
