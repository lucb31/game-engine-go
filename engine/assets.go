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
	PlainsTileset *ebiten.Image
	PlayerTileset *ebiten.Image
	TilesPerRow   int
	TileSize      int
}

func NewAssetManager() (*AssetManager, error) {
	// Load assets
	var err error
	im, err := ReadPngAsset("assets/player.png")
	if err != nil {
		fmt.Println("Could not read assets!", err.Error())
		return nil, err
	}
	playerImage := ebiten.NewImageFromImage(im)
	im, err = ReadPngAsset("assets/plains.png")
	if err != nil {
		fmt.Println("Could not read plains asset!", err.Error())
		return nil, err
	}
	plainsTileset := ebiten.NewImageFromImage(im)
	tileSize := 16
	tilesPerRow := int(plainsTileset.Bounds().Dx() / tileSize)
	return &AssetManager{PlayerTileset: playerImage, PlainsTileset: plainsTileset, TilesPerRow: tilesPerRow, TileSize: tileSize}, nil
}

func ReadPngAsset(path string) (image.Image, error) {
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
