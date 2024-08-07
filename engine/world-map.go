package engine

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

type MapTile int

const (
	Empty       MapTile = -1
	Gras        MapTile = 32
	Rock        MapTile = 42
	Undef       MapTile = 71
	mapTileSize         = 16
)

type WorldMap struct {
	tileset  Tileset
	tileData [][]MapTile
}

func (b *WorldMap) Draw(screen *ebiten.Image) {
	count := 0
	// Drawing WHOLE map. This is ok because there is no camera movement right now
	for row := range len(b.tileData) {
		for col := range len(b.tileData[0]) {
			mapTile := b.tileData[row][col]
			// Set tile position
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*mapTileSize), float64(row*mapTileSize))

			// If NOT undef tile, draw undef tile first to have it in the background
			if mapTile != Undef {
				subIm, err := b.tileset.GetTile(int(Undef))
				if err != nil {
					fmt.Println("Unable to draw background cell", err.Error())
					return
				}
				screen.DrawImage(subIm, &op)
			}
			if mapTile == Empty {
				continue
			}
			// Select correct tile from tileset
			subIm, err := b.tileset.GetTile(int(mapTile))
			if err != nil {
				fmt.Println("Unable to draw world map cell", err.Error())
				return
			}
			screen.DrawImage(subIm, &op)
			count++
		}
	}
}

// Generate new world map for widht & height dimensions IN PX
func NewWorldMap(width, height int64, path string, tileset Tileset) (*WorldMap, error) {
	// Read map data from provided path
	csvMapData, err := readCsvMap(path)
	if err != nil {
		return nil, err
	}
	mapData := make([][]MapTile, height/mapTileSize)
	// Copy map data in & fill remaining cells with placeholder tile
	for row := range height / mapTileSize {
		mapData[row] = make([]MapTile, width/mapTileSize)
		for col := range width / mapTileSize {
			if int64(len(csvMapData)) > row && int64(len(csvMapData[row])) > col {
				mapData[row][col] = csvMapData[row][col]
			} else {
				mapData[row][col] = Undef
			}
		}
	}
	return &WorldMap{tileData: mapData, tileset: tileset}, nil
}

func readCsvMap(path string) ([][]MapTile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
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
