package engine

import (
	"fmt"
	"log"

	"github.com/jakecoffman/cp"
)

type WorldMap interface {
	Draw(camera Camera)
	AddCsvLayer(mapCsv []byte, tileset *Tileset) error
}

type MultiLayerWorldMap struct {
	layers        []MapLayer
	width, height int64
}

// Creates a new multi layer world map
func NewMultiLayerWorldMap(width, height int64) (*MultiLayerWorldMap, error) {
	m := &MultiLayerWorldMap{width: width, height: height}
	return m, nil
}

// Draw all map layers
func (w *MultiLayerWorldMap) Draw(camera Camera) {
	if len(w.layers) == 0 {
		log.Println("Empty world map. No layers defined")
		return
	}
	for _, l := range w.layers {
		l.Draw(camera)
	}
}

func (w *MultiLayerWorldMap) AddSkyboxLayer(width, height int64, tileset *Tileset) error {
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

// Init & append layer from provided csv tile data and tileset
func (w *MultiLayerWorldMap) AddCsvLayer(mapCsv []byte, tileset *Tileset) error {
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

// Returns row, col of world position tile (rounded down)
func WorldPosToGridPos(pos cp.Vector) (int, int) {
	row := max(int(pos.Y/mapTileSize), 0)
	col := max(int(pos.X/mapTileSize), 0)
	return row, col
}
