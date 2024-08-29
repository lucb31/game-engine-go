package engine

import (
	"math"
	"math/rand"

	"github.com/jakecoffman/cp"
)

type HexSegment [][]MapTile

// Procedurally generated map based on hexagonal tiles
type HexWorldMap struct {
	*MultiLayerWorldMap

	groundLayer *BaseMapLayer
	propLayer   *BaseMapLayer
	// TODO: Datatype
	csvHexSegments []HexSegment
	center         cp.Vector
}

func NewProcHexWorldMap(width, height int64, center cp.Vector) (*HexWorldMap, error) {
	// Init base
	base, err := NewMultiLayerWorldMap(width, height)
	if err != nil {
		return nil, err
	}
	m := &HexWorldMap{MultiLayerWorldMap: base}
	m.center = center
	return m, nil
}

// Set up 2 empty layers with the provided tileset
// During level generation random hexagon map segments will be copied into these layers
// Hexagon map segments need to provide the same amount of layers
func (m *HexWorldMap) InitHexBaseLayers(tileset *Tileset) error {
	var err error
	// Init ground & prop layers
	groundLayer, err := NewEmptyMapLayer(m.width, m.height)
	if err != nil {
		return err
	}
	groundLayer.SetTileset(*tileset)
	m.groundLayer = groundLayer

	propLayer, err := NewEmptyMapLayer(m.width, m.height)
	if err != nil {
		return err
	}
	propLayer.SetTileset(*tileset)
	m.propLayer = propLayer

	m.layers = append(m.layers, m.groundLayer, m.propLayer)
	return nil
}

// Adds hexagon map segment csv data to the pool of available segments
// that the procedure will randomly choose from
func (m *HexWorldMap) AddHexSegment(mapCsv []byte) error {
	// Read map data from provided path
	csvMapData, err := ReadCsvFromBinary(mapCsv)
	if err != nil {
		return err
	}
	m.csvHexSegments = append(m.csvHexSegments, csvMapData)
	return nil
}

func (m *HexWorldMap) Generate() error {
	// FIX: Currently first hex segment is always used as start
	startingHex := m.csvHexSegments[0]
	// Temporary: Only work on ground layer
	// TODO: Add prop layer
	layer := m.groundLayer
	// Draw first hexagon at center position
	if err := layer.CopyMapDataToCenterPosition(startingHex, m.center); err != nil {
		return err
	}

	// Draw first ring of hexes
	return m.NewHexRing(m.center, m.csvHexSegments[0], layer)
}

func (m *HexWorldMap) getRandomSegment() HexSegment {
	idx := rand.Intn(len(m.csvHexSegments))
	return m.csvHexSegments[idx]
}

func (m *HexWorldMap) NewHexRing(center cp.Vector, csvMapData [][]MapTile, l *BaseMapLayer) error {
	// Determine hex radius from map data
	mapSizeX := len(csvMapData[0]) * mapTileSize
	// NOTE: This assumes that the map size along the X axis is twice the radius of the hexagon
	// and map size of all hex segments is equal
	radius := float64(mapSizeX / 2)
	// Ring of hexes by iterating over all 6 edges of the center hexagon
	// r = cos(30°)*R https://en.wikipedia.org/wiki/Hexagon#Parameters
	inradius := math.Cos(30.0/180.0*math.Pi) * radius
	steps := 6
	for i := 0; i < steps; i++ {
		// Calculate ring segment center position
		// Offset by 30°
		angle := (2.0*float64(i)/float64(steps) + 30.0/180.0) * math.Pi
		hexCenter := center.Add(cp.Vector{math.Cos(angle) * inradius * 2, math.Sin(angle) * inradius * 2})

		// Randomize which hexagon to pick, ideally flip and / or rotate
		mapData := m.getRandomSegment()

		// Copy to layer
		if err := l.CopyMapDataToCenterPosition(mapData, hexCenter); err != nil {
			return err
		}
	}
	return nil
}
