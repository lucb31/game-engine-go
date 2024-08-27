package engine

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type MapTile int
type WorldMapReader interface {
	TileAt(cp.Vector) (MapTile, error)
}

const (
	EmptyTile   MapTile = -1
	mapTileSize         = 16
)

type MapLayer interface {
	Draw(Camera) error
	// Return width & height of map layer
	Dimensions() (int, int)
	WorldMapReader
}

type BaseMapLayer struct {
	tileset  Tileset
	tileData [][]MapTile
}

// Generate new map layer for widht & height dimensions IN PX
func NewBaseMapLayer(width, height int64, mapCsv []byte, tileset *Tileset) (*BaseMapLayer, error) {
	// Read map data from provided path
	csvMapData, err := ReadCsvFromBinary(mapCsv)
	if err != nil {
		return nil, err
	}
	// +2 to add one additional tile if width / height mod tilesize is not 0
	cols := int64(width/mapTileSize) + 2
	rows := int64(height/mapTileSize) + 2
	mapData := make([][]MapTile, rows)
	// Copy map data in & fill remaining cells with empty tile (cannot keep at 0, because that already corresponds to a tile)
	for row := range rows {
		mapData[row] = make([]MapTile, cols)
		for col := range cols {
			if int64(len(csvMapData)) > row && int64(len(csvMapData[row])) > col {
				mapData[row][col] = csvMapData[row][col]
			} else {
				mapData[row][col] = EmptyTile
			}
		}
	}
	layer := &BaseMapLayer{tileData: mapData}
	if tileset != nil {
		layer.tileset = *tileset
	}
	return layer, nil
}

func NewEmptyMapLayer(width, height int64) (*BaseMapLayer, error) {
	// +2 to add one additional tile if width / height mod tilesize is not 0
	cols := int64(width/mapTileSize) + 2
	rows := int64(height/mapTileSize) + 2
	mapData := make([][]MapTile, rows)
	// Fill with Empty tile
	for row := range rows {
		mapData[row] = make([]MapTile, cols)
		for col := range cols {
			mapData[row][col] = EmptyTile
		}
	}
	layer := &BaseMapLayer{tileData: mapData}
	return layer, nil
}

func (l *BaseMapLayer) CopyMapDataToCenterPosition(csvMapData [][]MapTile, center cp.Vector) error {
	// Calc starting position (top left) of sub-layer from layer data bounds & center position
	mapSize := cp.Vector{float64(len(csvMapData[0]) * mapTileSize), float64(len(csvMapData) * mapTileSize)}
	dstTopLeft := center.Sub(mapSize.Mult(0.5))

	// Copy (in bounds) tiles from sub-layer to layer. Ignoring empty tiles
	offsetRow := int(dstTopLeft.Y / mapTileSize)
	offsetCol := int(dstTopLeft.X / mapTileSize)
	// offsetRow := int(math.Min(math.Max(dstTopLeft.Y/mapTileSize, 0), float64(len(l.tileData)-1)))
	// offsetCol := int(math.Min(math.Max(dstTopLeft.X/mapTileSize, 0), float64(len(l.tileData[0])-1)))
	for row := 0; row < len(csvMapData) && row+offsetRow < len(l.tileData); row++ {
		for col := 0; col < len(csvMapData[row]) && col+offsetCol < len(l.tileData[0]); col++ {
			// Skip out of bounds
			// FIX: This could be optimized by a better loop condition
			if row+offsetRow < 0 || col+offsetCol < 0 {
				continue
			}
			// Skip empty tiles
			if csvMapData[row][col] == EmptyTile {
				continue
			}
			l.tileData[row+offsetRow][col+offsetCol] = csvMapData[row][col]
		}
	}
	return nil
}

// Iterates over map layer data and draws tiles that are within viewport of camera
func (l *BaseMapLayer) Draw(camera Camera) error {
	topLeft, bottomRight := camera.Viewport()
	// We dont want to iterate over out of bounds rows and cols
	startingRow := int(math.Max(topLeft.Y/mapTileSize, 0))
	// + 1 to ensure that last row is included
	endRow := int(math.Min(bottomRight.Y/mapTileSize+1, float64(len(l.tileData))))
	startingCol := int(math.Max(topLeft.X/mapTileSize, 0))
	// + 1 to ensure that last col is included
	endCol := int(math.Min(bottomRight.X/mapTileSize+1, float64(len(l.tileData[0]))))

	for row := startingRow; row < endRow; row++ {
		for col := startingCol; col < endCol; col++ {
			mapTile := l.tileData[row][col]
			// Ignore empty cells
			if mapTile == EmptyTile {
				continue
			}

			// Set tile position
			op := ebiten.DrawImageOptions{}
			x, y := GridPosToTopLeftWorldPos(col, row)
			op.GeoM.Translate(x, y)
			// Select correct tile from tileset
			subIm, err := l.tileset.GetTile(int(mapTile))
			if err != nil {
				return fmt.Errorf("Unable to draw world map cell: %s", err.Error())
			}
			// Draw tile
			camera.DrawImage(subIm, &op)
		}
	}
	return nil
}

func (l *BaseMapLayer) TileAt(worldPos cp.Vector) (MapTile, error) {
	if worldPos.X < 0 || worldPos.Y < 0 {
		return EmptyTile, fmt.Errorf("Out of bounds")
	}
	row := int(worldPos.Y / mapTileSize)
	col := int(worldPos.X / mapTileSize)
	if len(l.tileData) <= row || len(l.tileData[0]) <= col {
		return EmptyTile, fmt.Errorf("Out of bounds")
	}
	return l.tileData[row][col], nil
}

func (l *BaseMapLayer) Dimensions() (int, int) {
	if len(l.tileData) == 0 {
		return 0, 0
	}
	return len(l.tileData), len(l.tileData[0])
}

func ReadCsvFromBinary(data []byte) ([][]MapTile, error) {
	reader := bytes.NewReader(data)
	return readCsv(reader)
}

func readCsvFromFile(path string) ([][]MapTile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readCsv(f)
}

func readCsv(r io.Reader) ([][]MapTile, error) {
	csvReader := csv.NewReader(r)
	mapData := [][]MapTile{}
	for row := 0; ; row++ {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		mapData = append(mapData, make([]MapTile, len(rec)))
		for col, data := range rec {
			intVal, err := strconv.Atoi(data)
			if err != nil {
				return nil, err
			}
			mapData[row][col] = MapTile(intVal)
		}
	}
	return mapData, nil
}
