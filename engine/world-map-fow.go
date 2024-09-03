package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type FogOfWar interface {
	Draw(camera Camera)
	DiscoverWithRadius(pos cp.Vector, innerRadius, outerRadius float64)
	VectorVisible(cp.Vector) bool
}

type DiscoveryLayer struct {
	discovered [][]uint8
	debugger   *ExecutionDebugger
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
	l.debugger = NewExecutionDebugger("FoW: Draw")
	return l, nil
}

// First implementation of fog of war rendering
// No optimizations. Draw every tile individually, use camera to offset to screen dimensions
// Avg execution time on initial view ~ 8.2ms
func (l *DiscoveryLayer) drawV0(camera Camera) {
	defer l.debugger.AvgExeTime()()
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

// Same row & col calculation, but
// Improvement: If multiple cols within one row have the same value, we combine the cols into a bigger segment
// Avg execution ~ 3.3 ms on initial screen, but even better the more (continuous) fog there is
// FIX: Weird effects on first col IF it is different than the second coll
// Weird effect on center of light radius
func (l *DiscoveryLayer) drawV1(camera Camera) {
	defer l.debugger.AvgExeTime()()
	topLeft, bottomRight := camera.Viewport()
	// We dont want to iterate over out of bounds rows and cols
	startingRow := int(math.Max(topLeft.Y/mapTileSize, 0))
	// + 1 to ensure that last row is included
	endRow := int(math.Min(bottomRight.Y/mapTileSize+1, float64(len(l.discovered))))
	startingCol := int(math.Max(topLeft.X/mapTileSize, 0))
	// + 1 to ensure that last col is included
	endCol := int(math.Min(bottomRight.X/mapTileSize+1, float64(len(l.discovered[0]))))

	for row := startingRow; row < endRow; row++ {
		prevAlpha := l.discovered[row][startingCol]
		prevCol := startingCol
		for col := startingCol; col < endCol; col++ {
			alpha := l.discovered[row][col]
			if alpha != prevAlpha {
				// Draw segment
				leftX, y := GridPosToTopLeftWorldPos(prevCol, row)
				topLeft := cp.Vector{leftX, y}
				// NOTE: +1 to span until next tile
				rightX, bottomY := GridPosToTopLeftWorldPos(col+1, row+1)
				bottomRight := cp.Vector{rightX, bottomY}
				camera.FillRect(topLeft, bottomRight, color.NRGBA{0, 0, 0, prevAlpha}, false)

				// Start new segment
				prevCol = col
				prevAlpha = alpha
			}
		}
		// Draw last segment (can be entire row)
		leftX, y := GridPosToTopLeftWorldPos(prevCol, row)
		topLeft := cp.Vector{leftX, y}
		rightX, bottomY := GridPosToTopLeftWorldPos(endCol+1, row+1)
		bottomRight := cp.Vector{rightX, bottomY}
		camera.FillRect(topLeft, bottomRight, color.NRGBA{0, 0, 0, prevAlpha}, false)
	}
}

// Ideas: Draw pixels with correct tilesize
// => a lot of blank space, shader will interpolate
// FIX: Unfinished. Couldnt get it to work
func (l *DiscoveryLayer) drawV2(camera Camera) {
	defer l.debugger.AvgExeTime()()

	screen := ebiten.NewImage(camera.ScreenWidth(), camera.ScreenHeight())
	px := make([]byte, 4*camera.ScreenWidth()*camera.ScreenHeight())

	maxRow := int(camera.ScreenHeight() / mapTileSize)
	maxCol := int(camera.ScreenWidth() / mapTileSize)
	camTopLeft, _ := camera.Viewport()
	for row := range maxRow {
		for col := range maxCol {
			// Discrete offset: Figure out tile to use
			tileCol := max(0, col-int(camTopLeft.X/mapTileSize))
			tileRow := max(0, row-int(camTopLeft.Y/mapTileSize))

			alpha := l.discovered[tileRow][tileCol]

			i := row*maxCol*mapTileSize*mapTileSize + col*mapTileSize
			px[4*i] = 0
			px[4*i+1] = 0
			px[4*i+2] = 0
			px[4*i+3] = alpha
		}
	}
	screen.WritePixels(px)
	camera.Screen().DrawImage(screen, &ebiten.DrawImageOptions{})
}

// Variant: Draw every tile as one pixel, then use scale transformation
// Incredibly fast! ~40 µs, but very clunky. Information loss in between map tile size
// Unfinished! Problems are
// 1. Out of bounds
// Fixes: Clunkyness by offseting / translating
func (l *DiscoveryLayer) drawV3(camera Camera) {
	defer l.debugger.AvgExeTime()()
	topLeft, bottomRight := camera.Viewport()
	// We dont want to iterate over out of bounds rows and cols
	startingRow := int(math.Max(topLeft.Y/mapTileSize, 0))
	// + 1 to ensure that last row is included
	endRow := int(math.Min(bottomRight.Y/mapTileSize+1, float64(len(l.discovered))))
	startingCol := int(math.Max(topLeft.X/mapTileSize, 0))
	// + 1 to ensure that last col is included
	endCol := int(math.Min(bottomRight.X/mapTileSize+1, float64(len(l.discovered[0]))))

	// Problem: with this apporach the dimension of the px array is changing if we're moving closer towards OOB
	// => px array dimension needs to be multiplicator of screen dimension
	// 	width := endCol - startingCol
	// 	height := endRow - startingRow

	// +2 because both starting and end tile might be a "fractional" tile
	width := int(camera.ScreenWidth()/mapTileSize) + 1
	height := int(camera.ScreenHeight()/mapTileSize) + 2
	px := make([]byte, 4*width*height)

	i := 0
	for row := startingRow; row < endRow; row++ {
		for col := startingCol; col < endCol; col++ {
			alpha := l.discovered[row][col]
			px[4*i] = 0
			px[4*i+1] = 0
			px[4*i+2] = 0
			px[4*i+3] = alpha
			i++
		}
	}
	image := ebiten.NewImage(width, height)
	image.WritePixels(px)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(mapTileSize, mapTileSize)
	offsetX := calcFloatingOffset(topLeft.X) * -1
	offsetY := calcFloatingOffset(topLeft.Y) * -1
	op.GeoM.Translate(offsetX, offsetY)
	camera.Screen().DrawImage(image, op)
}

// Same approach as in v3, but always draw same size array of screen dimensions
// Almost good. Correct display at ~50 µs avg render time
// FIX: Does not support zoom factors < 1.0, but that is acceptable for now
// Maybe fixed by using scaled map tile size everywhere. That would result in a bigger px array
func (l *DiscoveryLayer) Draw(camera Camera) {
	defer l.debugger.AvgExeTime()()
	topLeft, _ := camera.Viewport()

	// +2 because both starting and end tile might be a "fractional" tile
	width := int(camera.ScreenWidth()/mapTileSize) + 2
	height := int(camera.ScreenHeight()/mapTileSize) + 2
	px := make([]byte, 4*width*height)
	// Calculate discrete / tile offsets
	discreteOffsetX := int(topLeft.X / mapTileSize)
	discreteOffsetY := int(topLeft.Y / mapTileSize)

	// Write exactly one pixel with greyscale fog value
	i := 0
	for row := range height {
		for col := range width {
			// Apply discrete offset
			tileCol := col + discreteOffsetX
			tileRow := row + discreteOffsetY

			// Out of bounds
			if tileCol < 0 || tileCol > len(l.discovered[0])-1 || tileRow < 0 || tileRow > len(l.discovered)-1 {
				// Since we're going to disable the skybox we show maximum fog outside of bounds
				px[4*i+3] = fogMaxAlpha
			} else {
				alpha := l.discovered[tileRow][tileCol]
				px[4*i] = 0
				px[4*i+1] = 0
				px[4*i+2] = 0
				px[4*i+3] = alpha
			}
			i++
		}
	}
	image := ebiten.NewImage(width, height)
	image.WritePixels(px)

	op := &ebiten.DrawImageOptions{}
	// Scale by scaled map tile size
	// FIX: Does not work correctly for zoom levels < 1
	op.GeoM.Scale(mapTileSize*camera.Zoom(), mapTileSize*camera.Zoom())
	// Offset by fractional tile size AND one full tile size
	offsetX := calcFloatingOffset(topLeft.X)*-1 - mapTileSize
	offsetY := calcFloatingOffset(topLeft.Y)*-1 - mapTileSize
	op.GeoM.Translate(offsetX, offsetY)

	screen := ebiten.NewImage(camera.Screen().Bounds().Dx(), camera.Screen().Bounds().Dy())
	screen.DrawImage(image, op)
	// Draw
	camera.Screen().DrawImage(image, op)
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
