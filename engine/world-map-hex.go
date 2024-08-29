package engine

import (
	"log"
	"math"

	"github.com/jakecoffman/cp"
)

// Procedurally generated map based on hexagonal tiles
type HexWorldMap struct {
	*MultiLayerWorldMap
	groundLayer *BaseMapLayer
	propLayer   MapLayer
}

// TODO: Center argument currently unusde
func NewProcHexWorldMap(width, height int64, center cp.Vector) (*HexWorldMap, error) {
	// Init base
	base, err := NewMultiLayerWorldMap(width, height)
	if err != nil {
		return nil, err
	}
	m := &HexWorldMap{MultiLayerWorldMap: base}
	return m, nil
}

func (m *HexWorldMap) InitHexBaseLayers() error {
	var err error
	// Init ground & prop layers
	if m.groundLayer, err = NewEmptyMapLayer(m.width, m.height); err != nil {
		return err
	}
	if m.propLayer, err = NewEmptyMapLayer(m.width, m.height); err != nil {
		return err
	}
	m.layers = append(m.layers, m.groundLayer, m.propLayer)
	return nil
}

// TODO: Dont need to return layer
// TODO: Tileset needs to be the same for all. Should be on world map level, not layer
func (m *HexWorldMap) NewHexLayer(center cp.Vector, mapCsv []byte, tileset *Tileset) (*BaseMapLayer, error) {
	// Temporary: Only work on ground layer
	// TODO: Add prop layer
	// TODO: Split center tile from radius tiles
	l := m.groundLayer

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
	log.Println("succes")
	return l, nil
}

// func (w *HexWorldMap) AddHexLayer(center cp.Vector, mapCsv []byte, tileset *Tileset) error {
// 	l, err := NewHexLayer(w.width, w.height, center, mapCsv, tileset)
// 	if err != nil {
// 		return err
// 	}
// 	w.layers = append(w.layers, l)
// 	return nil
// }
