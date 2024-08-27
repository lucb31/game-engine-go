package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
)

func NewHexLayer(width, height int64, center cp.Vector, mapCsv []byte, tileset *Tileset) (*BaseMapLayer, error) {
	// Initialize empty base layer. We will then copy hexagon tiles onto it
	l, err := NewEmptyMapLayer(width, height)
	if err != nil {
		return nil, err
	}

	// Read map data from provided path
	csvMapData, err := ReadCsvFromBinary(mapCsv)
	if err != nil {
		return nil, err
	}
	// Determine hex radius from map data
	mapSizeX := len(csvMapData[0]) * mapTileSize
	// NOTE: This assumes that the map size along the X axis is twice the radius of the hexagon
	radius := float64(mapSizeX / 2)

	// Draw first hexagon at center position
	if err := l.CopyMapDataToCenterPosition(csvMapData, center); err != nil {
		return nil, err
	}

	// Second ring of hexes with new center. By iterating over all 6 edges of the center hexagon
	// r = cos(30°)*R https://en.wikipedia.org/wiki/Hexagon#Parameters
	inradius := math.Cos(30.0/180.0*math.Pi) * radius
	steps := 6
	for i := 0; i < steps; i++ {
		// Offset by 30°
		angle := (2.0*float64(i)/float64(steps) + 30.0/180.0) * math.Pi
		hexCenter := center.Add(cp.Vector{math.Cos(angle) * inradius * 2, math.Sin(angle) * inradius * 2})
		if err := l.CopyMapDataToCenterPosition(csvMapData, hexCenter); err != nil {
			return nil, err
		}
	}
	l.tileset = *tileset
	return l, nil
}

func strokeHex(screen *ebiten.Image, center cp.Vector, radius float64, stroke float32, color color.Color) {
	verts := []cp.Vector{}
	for i := 0; i < 6; i++ {
		angle := 2.0 * math.Pi / 6.0 * float64(i)
		verts = append(verts, center.Add(cp.Vector{math.Cos(angle) * float64(radius), math.Sin(angle) * float64(radius)}))
		if i > 0 {
			vector.StrokeLine(screen, float32(verts[i-1].X), float32(verts[i-1].Y), float32(verts[i].X), float32(verts[i].Y), stroke, color, false)
		}
	}
	// Last line segment
	vector.StrokeLine(screen, float32(verts[0].X), float32(verts[0].Y), float32(verts[5].X), float32(verts[5].Y), stroke, color, false)
}

func drawHexes(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	stroke := float32(2.0)
	blue := color.NRGBA{0, 0, 255, 255}
	radius := 64.0
	inradius := math.Cos(30.0/180.0*math.Pi) * radius
	origin := cp.Vector{float64(screenWidth) / 2, float64(screenHeight) / 2}
	cx := float32(screenWidth / 2)
	cy := float32(screenHeight / 2)
	// Inner circle around starting point
	vector.StrokeCircle(screen, cx, cy, float32(radius), stroke, color.White, false)
	// Outer circle around starting point
	vector.StrokeCircle(screen, cx, cy, float32(inradius)*2, stroke, color.White, false)

	// First hex around center
	strokeHex(screen, origin, radius, stroke, blue)

	// Second ring of hexes with new center
	for i := 0; i < 6; i++ {
		// Offset by 30°
		angle := (2.0/6.0*float64(i) + 30.0/180.0) * math.Pi
		center := origin.Add(cp.Vector{math.Cos(angle) * inradius * 2, math.Sin(angle) * inradius * 2})
		strokeHex(screen, center, radius, stroke, color.NRGBA{255, 0, 0, 255})
	}
}
