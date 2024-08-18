package engine

import (
	"github.com/jakecoffman/cp"
)

type WorldMap struct {
	layers        []*MapLayer
	width, height int64
}

// Creates a new multi layer world map. Need to provide data for base layer
// since map cannot have 0 layers
func NewWorldMap(width, height int64, mapCsv []byte, tileset *Tileset) (*WorldMap, error) {
	m := &WorldMap{width: width, height: height}
	if err := m.AddLayer(mapCsv, tileset); err != nil {
		return nil, err
	}
	return m, nil
}

// Draw all map layers
func (m *WorldMap) Draw(camera Camera) {
	for _, l := range m.layers {
		l.Draw(camera)
	}
}

func (w *WorldMap) TileAt(worldPos cp.Vector) (MapTile, error) {
	return w.layers[0].TileAt(worldPos)
}

func (w *WorldMap) AddLayer(mapCsv []byte, tileset *Tileset) error {
	l, err := NewMapLayer(w.width, w.height, mapCsv, tileset)
	if err != nil {
		return err
	}
	w.layers = append(w.layers, l)
	return nil
}

// Returns vector centered on grid
func SnapToGrid(v cp.Vector, gridX int, gridY int) cp.Vector {
	return cp.Vector{X: float64(int(v.X/float64(gridX))*gridX) + float64(gridX)/2, Y: float64(int(v.Y/float64(gridY))*gridY) + float64(gridY/2)}
}
