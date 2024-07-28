package engine

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type tileResource struct {
	Key      string
	Path     string
	TileSize int
}

var resources []tileResource = []tileResource{
	{"player", "assets/player.png", 48},
	{"plains", "assets/plains.png", 16},
	{"fences", "assets/fences.png", 16},
}

type Tileset struct {
	Image       *ebiten.Image
	TilesPerRow int
	TileSize    int
}

type AssetManager struct {
	Tilesets map[string]Tileset
}

func NewAssetManager() (*AssetManager, error) {
	// Load all tilesets specified in 'resources' into tileset map
	tiles := map[string]Tileset{}
	for _, res := range resources {
		tileset, err := loadTileset(res.Path, res.TileSize)
		if err != nil {
			return nil, err
		}
		tiles[res.Key] = *tileset
	}

	return &AssetManager{Tilesets: tiles}, nil
}

func (a *AssetManager) GetTile(tileSetKey string, tileIdx int) (*ebiten.Image, error) {
	tileSet, ok := a.Tilesets[tileSetKey]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown tileset %v", tileSetKey)
	}
	tileX := tileIdx % tileSet.TilesPerRow
	tileY := int(tileIdx / tileSet.TilesPerRow)
	// Selecting sub image based on tile information
	return tileSet.Image.SubImage(image.Rect(
		tileX*tileSet.TileSize,
		tileY*tileSet.TileSize,
		(tileX+1)*tileSet.TileSize,
		(tileY+1)*tileSet.TileSize,
	)).(*ebiten.Image), nil
}

func loadTileset(path string, tileSize int) (*Tileset, error) {
	im, err := readPngAsset(path)
	if err != nil {
		fmt.Println("Could not read assets!", err.Error())
		return nil, err
	}
	return &Tileset{
		Image:       ebiten.NewImageFromImage(im),
		TilesPerRow: int(im.Bounds().Dx() / tileSize),
		TileSize:    tileSize,
	}, nil
}

func readPngAsset(path string) (image.Image, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewReader(dat))
	if err != nil {
		return nil, err
	}
	return im, nil
}

type GameEntity interface {
	Draw(*ebiten.Image)
	Update()
}

type GameObj struct {
	asset           *GameAsset
	activeAnimation string
	posX            int
	posY            int
}

type GameAsset struct {
	Name        string
	Animations  map[string]GameAssetAnimation
	frameHeight int8
	frameWidht  int8
}

type GameAssetAnimation struct {
	Name string
	// Number of frames to play
	FrameCount int
	// Starting position in the TileMap
	TileIdx int
}
