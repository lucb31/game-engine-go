package engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/jakecoffman/cp"
)

type HexSegmentPattern [][]MapTile

// Procedurally generated map based on hexagonal tiles
type HexWorldMap struct {
	*MultiLayerWorldMap

	groundLayer *BaseMapLayer
	propLayer   *BaseMapLayer
	// TODO: Datatype
	// Pool of segments to randomly choose from
	segmentPool []HexSegmentPattern

	// Center pos of initial hex
	center cp.Vector
	// Outer radius of all hexes
	radius float64
	// Inner radius of all hexes
	inradius float64
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

	// Determine hex radius from map data
	mapSizeX := len(csvMapData[0]) * mapTileSize
	// NOTE: This assumes that the map size along the X axis is twice the radius of the hexagon
	radius := float64(mapSizeX / 2)
	// Ensure radius of all hex segments is equal
	if m.radius == 0 {
		m.SetRadius(radius)
	} else if m.radius != radius {
		return fmt.Errorf("Hex segment radius does not match.")
	}

	// Add to pool
	m.segmentPool = append(m.segmentPool, csvMapData)
	return nil
}

func (m *HexWorldMap) SetRadius(radius float64) {
	m.radius = radius
	// r = cos(30°)*R https://en.wikipedia.org/wiki/Hexagon#Parameters
	m.inradius = math.Cos(30.0/180.0*math.Pi) * radius
}
func (m *HexWorldMap) Radius() float64 { return m.radius }

func (m *HexWorldMap) Generate() error {
	// FIX: Currently first hex segment is always used as start
	startingHex := m.segmentPool[0]
	// Temporary: Only work on ground layer
	// TODO: Add prop layer
	layer := m.groundLayer
	// Draw first hexagon at center position
	if err := layer.CopyMapDataToCenterPosition(startingHex, m.center); err != nil {
		return err
	}

	// Draw first ring of hexes
	return m.NewHexRing(m.center, m.segmentPool[0], layer)
}

func (m *HexWorldMap) getRandomSegment() HexSegmentPattern {
	idx := rand.Intn(len(m.segmentPool))
	return m.segmentPool[idx]
}

func (m *HexWorldMap) NewHexRing(center cp.Vector, csvMapData [][]MapTile, l *BaseMapLayer) error {
	// Ring of hexes by iterating over all 6 edges of the center hexagon
	steps := 6
	for i := 0; i < steps; i++ {
		// Calculate ring segment center position
		// Offset by 30°
		angle := (2.0*float64(i)/float64(steps) + 30.0/180.0) * math.Pi
		hexCenter := center.Add(cp.Vector{math.Cos(angle) * m.inradius * 2, math.Sin(angle) * m.inradius * 2})

		// Randomize which hexagon to pick, ideally flip and / or rotate
		mapData := m.getRandomSegment()

		// Copy to layer
		if err := l.CopyMapDataToCenterPosition(mapData, hexCenter); err != nil {
			return err
		}
	}
	return nil
}
