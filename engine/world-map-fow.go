package engine

import (
	"image/color"
	"math"

	"github.com/jakecoffman/cp"
)

type FogOfWar interface {
	Draw(camera Camera)
	DiscoverWithRadius(pos cp.Vector, innerRadius, outerRadius float64)
	VectorVisible(cp.Vector) bool
}

type DiscoveryLayer struct {
	discovered [][]uint8
}

const (
	// Alpha value used for maximum fog
	fogMaxAlpha = uint8(250)
)

func NewDiscoveryLayer(width, height int64) (*DiscoveryLayer, error) {
	// +2 to add one additional tile if width / height mod tilesize is not 0
	cols := int64(width/mapTileSize) + 2
	rows := int64(height/mapTileSize) + 2
	mapData := make([][]uint8, rows)
	// Fill with default alpha value
	for row := range rows {
		mapData[row] = make([]uint8, cols)
		for col := range cols {
			mapData[row][col] = fogMaxAlpha
		}
	}
	l := &DiscoveryLayer{discovered: mapData}
	return l, nil
}

func (l *DiscoveryLayer) Draw(camera Camera) {
	topLeft, bottomRight := camera.Viewport()
	// We dont want to iterate over out of bounds rows and cols
	startingRow := int(math.Max(topLeft.Y/mapTileSize, 0))
	// + 1 to ensure that last row is included
	endRow := int(math.Min(bottomRight.Y/mapTileSize+1, float64(len(l.discovered))))
	startingCol := int(math.Max(topLeft.X/mapTileSize, 0))
	// + 1 to ensure that last col is included
	endCol := int(math.Min(bottomRight.X/mapTileSize+1, float64(len(l.discovered[0]))))

	for row := startingRow; row < endRow; row++ {
		for col := startingCol; col < endCol; col++ {
			mapTile := l.discovered[row][col]

			// Set tile position
			x, y := GridPosToTopLeftWorldPos(col, row)
			topLeft := cp.Vector{x, y}
			bottomRight := topLeft.Add(cp.Vector{mapTileSize, mapTileSize})
			camera.FillRect(topLeft, bottomRight, color.NRGBA{0, 0, 0, mapTile}, false)
		}
	}
}

func (l *DiscoveryLayer) VectorVisible(vec cp.Vector) bool {
	row, col := WorldPosToGridPos(vec)
	return l.discovered[row][col] < fogMaxAlpha
}

func (l *DiscoveryLayer) DiscoverWithRadius(pos cp.Vector, innerRadius float64, outerRadius float64) {
	radius := int(outerRadius / mapTileSize)

	// Cal starting row & max values
	rowPos, colPos := WorldPosToGridPos(pos)
	row := max(0, rowPos-radius)
	maxRow := min(rowPos+radius+1, len(l.discovered))
	maxCol := min(colPos+radius+1, len(l.discovered[0]))

	for ; row < maxRow; row++ {
		col := max(0, colPos-radius)
		for ; col < maxCol; col++ {
			dRow := rowPos - row
			dCol := colPos - col
			newGradient := calcGradient(dRow, dCol, innerRadius, outerRadius)
			l.discovered[row][col] = min(l.discovered[row][col], newGradient)
		}
	}
}

func calcGradient(dRow, dCol int, innerRadius, outerRadius float64) uint8 {
	fullVisibilityLengthSq := (innerRadius / mapTileSize) * (innerRadius / mapTileSize)
	noVisibilityLengthSq := (outerRadius / mapTileSize) * (outerRadius / mapTileSize)
	lengthSq := dRow*dRow + dCol*dCol

	x := (float64(lengthSq) - fullVisibilityLengthSq) / (noVisibilityLengthSq - fullVisibilityLengthSq)
	x = cp.Clamp01(x)
	// Linear gradient
	return uint8(x * float64(fogMaxAlpha))
}
