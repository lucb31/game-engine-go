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
	objects      []*GameObj
	player       GameEntity
	Biome        [][]Biome
	Width        int64
	Height       int64
	FrameCount   int64
	AssetManager *AssetManager
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	w.drawBiomes(screen)
	w.player.Draw(screen)
}

func (w *GameWorld) Update() {
	w.FrameCount++
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

func NewWorld(width int64, height int64) (*GameWorld, error) {
	am, err := NewAssetManager()
	if err != nil {
		return nil, err
	}
	biome, err := createBiome(width, height)
	if err != nil {
		return nil, err
	}
	w := GameWorld{Biome: biome, Width: width, Height: height, AssetManager: am}
	player, err := NewPlayer(&w)
	if err != nil {
		return &w, nil
	}
	w.player = player
	return &w, nil
}
