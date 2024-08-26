package engine

import (
	"fmt"
	"log"

	"github.com/jakecoffman/cp"
)

type WorldMap struct {
	layers        []MapLayer
	width, height int64
}

// Creates a new multi layer world map
func NewWorldMap(width, height int64) (*WorldMap, error) {
	m := &WorldMap{width: width, height: height}
	return m, nil
}

// Draw all map layers
func (w *WorldMap) Draw(camera Camera) {
	if len(w.layers) == 0 {
		log.Println("Empty world map. No layers defined")
		return
	}
	for _, l := range w.layers {
		l.Draw(camera)
	}
}

func (w *WorldMap) TileAt(worldPos cp.Vector) (MapTile, error) {
	return w.layers[0].TileAt(worldPos)
}

func (w *WorldMap) AddSkyboxLayer(width, height int64, tileset *Tileset) error {
	if len(w.layers) > 0 {
		return fmt.Errorf("Map already has existing layers. Skybox needs to be added as first layer")
	}
	l, err := NewSkyboxLayer(width, height, tileset)
	if err != nil {
		return err
	}
	w.layers = append(w.layers, l)
	return nil
}

func (w *WorldMap) AddLayer(mapCsv []byte, tileset *Tileset) error {
	l, err := NewBaseMapLayer(w.width, w.height, mapCsv, tileset)
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
