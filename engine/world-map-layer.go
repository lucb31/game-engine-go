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
	Empty       MapTile = -1
	Dirt        MapTile = 8
	Undef       MapTile = 71
	mapTileSize         = 16
)

type MapLayer struct {
	tileset  Tileset
	tileData [][]MapTile
}

// Generate new map layer for widht & height dimensions IN PX
func NewMapLayer(width, height int64, mapCsv []byte, tileset *Tileset) (*MapLayer, error) {
	// Read map data from provided path
	csvMapData, err := readCsvFromBinary(mapCsv)
	if err != nil {
		return nil, err
	}
	// +2 to add one additional tile if width / height mod tilesize is not 0
	cols := int64(width/mapTileSize) + 2
	rows := int64(height/mapTileSize) + 2
	mapData := make([][]MapTile, rows)
	// Copy map data in & fill remaining cells with placeholder tile
	for row := range rows {
		mapData[row] = make([]MapTile, cols)
		for col := range cols {
			if int64(len(csvMapData)) > row && int64(len(csvMapData[row])) > col {
				mapData[row][col] = csvMapData[row][col]
			} else {
				mapData[row][col] = Undef
			}
		}
	}
	return &MapLayer{tileData: mapData, tileset: *tileset}, nil
}

func (l *MapLayer) Draw(camera Camera) {
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
			// Set tile position
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*mapTileSize), float64(row*mapTileSize))

			// If NOT undef tile, draw dirt tile first to have it in the background
			if mapTile != Undef {
				subIm, err := l.tileset.GetTile(int(Dirt))
				if err != nil {
					fmt.Println("Unable to draw background cell", err.Error())
					return
				}
				camera.DrawImage(subIm, &op)
			}
			if mapTile == Empty {
				continue
			}
			// Select correct tile from tileset
			subIm, err := l.tileset.GetTile(int(mapTile))
			if err != nil {
				fmt.Println("Unable to draw world map cell", err.Error())
				return
			}
			camera.DrawImage(subIm, &op)
		}
	}
}

// NOTE: Currently only required to determine buildable tiles
func (l *MapLayer) TileAt(worldPos cp.Vector) (MapTile, error) {
	if worldPos.X < 0 || worldPos.Y < 0 {
		return Undef, fmt.Errorf("Out of bounds")
	}
	row := int(worldPos.Y / mapTileSize)
	col := int(worldPos.X / mapTileSize)
	if len(l.tileData) <= row || len(l.tileData[0]) <= col {
		return Undef, fmt.Errorf("Out of bounds")
	}
	return l.tileData[row][col], nil
}

func readCsvFromBinary(data []byte) ([][]MapTile, error) {
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
