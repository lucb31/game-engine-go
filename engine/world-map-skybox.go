package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type SkyboxLayer struct {
	lastAnimationTick time.Time
	width, height     int64

	// duplicate with base layer
	tileset  Tileset
	tileData [][]MapTile
}

const skyboxSpeed = 0

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
			tileData[row][col] = placeholderTile()
		}
	}
	l.tileData = tileData
}

func (l *SkyboxLayer) TileAt(cp.Vector) (MapTile, error) {
	return 0, fmt.Errorf("Missing implementation")
}

func (l *SkyboxLayer) Draw(cam Camera) error {
	// Animation
	// Check if we need to move the skybox
	// TODO: Deprecate & add parallax effect instead
	if skyboxSpeed > 0 {
		now := time.Now()
		diff := now.Sub(l.lastAnimationTick)
		if diff.Seconds() > skyboxSpeed {
			// Next tick
			l.Seed()
			l.lastAnimationTick = now
		}
	}

	for row, rowData := range l.tileData {
		for col, mapTile := range rowData {
			// Set tile position
			op := ebiten.DrawImageOptions{}
			x, y := GridPosToTopLeftWorldPos(col, row)
			op.GeoM.Translate(x, y)
			// Select correct tile from tileset
			subIm, err := l.tileset.GetTile(int(mapTile))
			if err != nil {
				return fmt.Errorf("Unable to draw world map cell", err.Error())
			}
			// Draw tile TO SCREEN, not using camera offset
			cam.Screen().DrawImage(subIm, &op)
		}
	}

	return nil
}
func placeholderTile() MapTile {
	if rand.Intn(30) < 29 {
		return MapTile(467)
	}
	// Random star
	// 436 - 443
	return MapTile(rand.Intn(8) + 436)
}
