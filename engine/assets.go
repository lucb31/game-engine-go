package engine

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type AssetManager struct {
	Tilesets        map[string]Tileset
	CharacterAssets map[string]CharacterAsset
}

func NewAssetManager() (*AssetManager, error) {
	am := &AssetManager{}
	var err error

	am.Tilesets, err = loadEnvironmentTilesets()
	if err != nil {
		return nil, err
	}

	am.CharacterAssets, err = loadCharacterAssets()
	if err != nil {
		return nil, err
	}

	return am, nil
}

func (a *AssetManager) GetTile(tileSetKey string, tileIdx int) (*ebiten.Image, error) {
	tileSet, ok := a.Tilesets[tileSetKey]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown tileset %v", tileSetKey)
	}
	return tileSet.GetTile(tileIdx)
}

type tileResource struct {
	Key      string
	Path     string
	TileSize int
}

// TODO: Load these from config file
var tileResources []tileResource = []tileResource{
	{"plains", "assets/plains.png", 16},
	{"fences", "assets/fences.png", 16},
}

// Load tilesets for static resources
func loadEnvironmentTilesets() (map[string]Tileset, error) {
	tiles := map[string]Tileset{}
	for _, res := range tileResources {
		tileset, err := loadTileset(res.Path, res.TileSize)
		if err != nil {
			return nil, err
		}
		tiles[res.Key] = *tileset
	}
	return tiles, nil
}

type characterResource struct {
	Key        string
	Path       string
	TileSize   int
	Animations map[string]GameAssetAnimation
}

// TODO: Load these from config file
var characterResources []characterResource = []characterResource{
	{
		"player",
		"assets/player.png",
		48,
		map[string]GameAssetAnimation{
			"walk_horizontal": {StartTile: 24, FrameCount: 6},
			"walk_north":      {StartTile: 30, FrameCount: 6},
			"walk_south":      {StartTile: 18, FrameCount: 6},
		},
	},
}

// Load characters
func loadCharacterAssets() (map[string]CharacterAsset, error) {
	characters := map[string]CharacterAsset{}
	for _, res := range characterResources {
		tileset, err := loadTileset(res.Path, res.TileSize)
		if err != nil {
			return nil, err
		}
		asset := CharacterAsset{
			Tileset:        *tileset,
			Animations:     res.Animations,
			animationSpeed: 6,
		}
		characters[res.Key] = asset
	}
	return characters, nil
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
