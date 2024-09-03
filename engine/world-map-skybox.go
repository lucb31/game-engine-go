package engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type SkyboxLayer struct {
	width, height int64

	// duplicate with base layer
	tileset  Tileset
	tileData [][]MapTile
}

const parallaxSpeed = -0.2

// Dimension should equal camera viewport
func NewSkyboxLayer(width, height int64, tileset *Tileset) (*SkyboxLayer, error) {
	layer := &SkyboxLayer{}
	if tileset == nil {
		return nil, fmt.Errorf("Missing tileset")
	}
	layer.tileset = *tileset
	layer.width = width
	layer.height = height
	layer.Seed()

	return layer, nil
}

func (l *SkyboxLayer) Seed() {
	// Seed
	// TODO: Question the +2
	// +2 to add one additional tile if width / height mod tilesize is not 0
	cols := int64(l.width/mapTileSize) + 2
	rows := int64(l.height/mapTileSize) + 2
	tileData := make([][]MapTile, rows)
	for row := range rows {
		tileData[row] = make([]MapTile, cols)
		for col := range cols {
			tileData[row][col] = randomStarTile()
		}
	}
	l.tileData = tileData
}

func (l *SkyboxLayer) TileAt(cp.Vector) (MapTile, error) {
	return 0, fmt.Errorf("Missing implementation")
}

func (l *SkyboxLayer) Draw(cam Camera) error {
	camTopLeft, _ := cam.Viewport()
	camPosWithParallaxFactor := camTopLeft.Mult(parallaxSpeed)
	// +2 because we need an extra tile at start and beginning to account for fraction tiles
	// NOTE: Using screen dimensions here,not viewport. We dont need to scale the skybox with zoom factor
	for row := range int(cam.ScreenHeight()/mapTileSize) + 2 {
		for col := range int(cam.ScreenWidth()/mapTileSize) + 2 {
			// Discrete offset: Figure out tile to use
			tileCol := calcDiscreteOffset(col, camPosWithParallaxFactor.X, len(l.tileData[0]))
			tileRow := calcDiscreteOffset(row, camPosWithParallaxFactor.Y, len(l.tileData))
			mapTile := l.tileData[tileRow][tileCol]

			// Add floating offsets
			x, y := GridPosToTopLeftWorldPos(col, row)
			x += calcFloatingOffset(camPosWithParallaxFactor.X)
			y += calcFloatingOffset(camPosWithParallaxFactor.Y)

			// Select correct tile from tileset
			subIm, err := l.tileset.GetTile(int(mapTile))
			if err != nil {
				return fmt.Errorf("Unable to draw world map cell: %s", err.Error())
			}
			// Draw
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, y)
			// Draw tile TO SCREEN, not using camera offset, because thats already accounted by discreate
			// & floating offset
			cam.Screen().DrawImage(subIm, &op)
		}
	}
	return nil
}

func (l *SkyboxLayer) Dimensions() (int, int) {
	if len(l.tileData) == 0 {
		return 0, 0
	}
	return len(l.tileData), len(l.tileData[0])
}

func calcDiscreteOffset(intVal int, floatVal float64, maxVal int) int {
	lastCol := maxVal - 1
	tileCol := (intVal - int(floatVal/mapTileSize)) % lastCol
	if tileCol < 0 {
		tileCol = lastCol + tileCol
	}
	return tileCol
}
func calcFloatingOffset(v float64) float64 {
	intVal, floatVal := math.Modf(v)
	// This ensures that the first tile is always visible
	// Without, there will be a gap in the top left screen
	offset := float64(-mapTileSize)
	offset += float64(int(intVal)%mapTileSize) + floatVal
	return offset
}

func randomStarTile() MapTile {
	if rand.Intn(30) < 29 {
		return MapTile(467)
	}
	// Random star from 4x8 tileset starting at idx 436 with rowsize 29
	col := rand.Intn(8)
	row := rand.Intn(4)
	tileIdx := col + row*29 + 436
	return MapTile(tileIdx)
}
