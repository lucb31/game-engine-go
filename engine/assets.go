package engine

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucb31/game-engine-go/bin/assets"
)

type AssetManager interface {
	CharacterAsset(string) (*CharacterAsset, error)
	ProjectileAsset(string) (*ProjectileAsset, error)
	Tileset(string) (*Tileset, error)
}

type AssetManagerImpl struct {
	Tilesets         map[string]Tileset
	characterAssets  map[string]CharacterAsset
	projectileAssets map[string]ProjectileAsset
}

func NewAssetManager(frameCount *int64) (*AssetManagerImpl, error) {
	am := &AssetManagerImpl{}
	var err error

	am.Tilesets, err = loadEnvironmentTilesets()
	if err != nil {
		return nil, err
	}

	am.characterAssets, err = loadCharacterAssets(frameCount)
	if err != nil {
		return nil, err
	}

	am.projectileAssets, err = loadProjectileAssets(frameCount)
	if err != nil {
		return nil, err
	}

	return am, nil
}

func (a *AssetManagerImpl) GetTile(tileSetKey string, tileIdx int) (*ebiten.Image, error) {
	tileSet, ok := a.Tilesets[tileSetKey]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown tileset %v", tileSetKey)
	}
	return tileSet.GetTile(tileIdx)
}

func (a *AssetManagerImpl) CharacterAsset(identifier string) (*CharacterAsset, error) {
	res, ok := a.characterAssets[identifier]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown asset %s", identifier)
	}
	return &res, nil
}

func (a *AssetManagerImpl) ProjectileAsset(identifier string) (*ProjectileAsset, error) {
	res, ok := a.projectileAssets[identifier]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown asset %s", identifier)
	}
	return &res, nil
}

func (a *AssetManagerImpl) Tileset(identifier string) (*Tileset, error) {
	res, ok := a.Tilesets[identifier]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown asset %s", identifier)
	}
	return &res, nil
}

type tileResource struct {
	Key       string
	ImageData []byte
	TileSize  int
}

// TODO: Load these from config file
var tileResources []tileResource = []tileResource{
	{"plains", assets.Plains, 16},
	{"grounds", assets.Grounds, 16},
	{"fences", assets.Fences, 16},
}

// Load tilesets for static resources
func loadEnvironmentTilesets() (map[string]Tileset, error) {
	tiles := map[string]Tileset{}
	for _, res := range tileResources {
		tileset, err := loadTileset(res.ImageData, res.TileSize, 1.0)
		if err != nil {
			return nil, err
		}
		tiles[res.Key] = *tileset
	}
	return tiles, nil
}

type characterResource struct {
	Key        string
	ImageData  []byte
	TileSize   int
	Animations map[string]GameAssetAnimation
	OffsetX    float64
	OffsetY    float64
	Scale      float64
}

// TODO: Load these from config file
var characterResources []characterResource = []characterResource{
	{
		"player",
		assets.Player,
		48,
		map[string]GameAssetAnimation{
			"walk_east": {StartTile: 24, FrameCount: 6},
			"walk_west": {StartTile: 24, FrameCount: 6, Flip: true},
			"idle_east": {StartTile: 6, FrameCount: 6},
			"idle_west": {StartTile: 6, FrameCount: 6, Flip: true},
		},
		-44,
		-60,
		1.8,
	},
	{
		"npc-torch",
		assets.NpcTorch,
		192,
		map[string]GameAssetAnimation{
			"walk_east": {StartTile: 6, FrameCount: 6},
			"walk_west": {StartTile: 6, FrameCount: 6, Flip: true},
			"idle_east": {StartTile: 0, FrameCount: 6},
			"idle_west": {StartTile: 0, FrameCount: 6, Flip: true},
		},
		-40,
		-37,
		0.4,
	},
	{
		"npc-orc",
		assets.Orc,
		100,
		map[string]GameAssetAnimation{
			"walk_east": {StartTile: 8, FrameCount: 6},
			"walk_west": {StartTile: 8, FrameCount: 6, Flip: true},
			"idle_east": {StartTile: 0, FrameCount: 6},
			"idle_west": {StartTile: 0, FrameCount: 6, Flip: true},
		},
		-60,
		-60,
		1.2,
	},
	{
		"npc-slime",
		assets.Slime,
		24,
		map[string]GameAssetAnimation{
			"walk_east": {StartTile: 0, FrameCount: 2},
			"walk_west": {StartTile: 0, FrameCount: 2, Flip: true},
			"idle_east": {StartTile: 0, FrameCount: 2},
			"idle_west": {StartTile: 0, FrameCount: 2, Flip: true},
		},
		-16,
		-16,
		1.5,
	},
	{
		"tower-blue",
		assets.TowerBlue,
		256,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 4},
		},
		-28,
		-20,
		0.22,
	},
	{
		"tower-red",
		assets.TowerRed,
		256,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1},
		},
		-28,
		-30,
		0.22,
	},
	{
		"castle",
		assets.Castle,
		192,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1},
		},
		-44,
		-28,
		0.44,
	},
}

// Load characters
func loadCharacterAssets(frameCount *int64) (map[string]CharacterAsset, error) {
	characters := map[string]CharacterAsset{}
	for _, res := range characterResources {
		tileset, err := loadTileset(res.ImageData, res.TileSize, res.Scale)
		if err != nil {
			return nil, err
		}
		asset := CharacterAsset{
			Tileset:        *tileset,
			Animations:     res.Animations,
			animationSpeed: 6,
			offsetX:        res.OffsetX,
			offsetY:        res.OffsetY,
			currentFrame:   frameCount,
		}
		characters[res.Key] = asset
	}
	return characters, nil
}

func loadProjectileAssets(frameCount *int64) (map[string]ProjectileAsset, error) {
	projectiles := map[string]ProjectileAsset{}

	// Load bone
	im, err := loadImageFromBinaryPng(assets.Bone)
	if err != nil {
		return nil, err
	}
	// Scale image to target size 16x61
	targetSize := 16
	rawIm := ebiten.NewImageFromImage(im)
	scaledIm := ebiten.NewImage(targetSize, targetSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(targetSize)/float64(rawIm.Bounds().Dx()), float64(targetSize)/float64(rawIm.Bounds().Dy()))
	scaledIm.DrawImage(rawIm, op)
	// Add to asset map
	projectiles["bone"] = ProjectileAsset{
		Image:          scaledIm,
		currentFrame:   frameCount,
		animationSpeed: 2,
	}

	// Load arrow
	im, err = loadImageFromBinaryPng(assets.Arrow)
	if err != nil {
		return nil, err
	}
	projectiles["arrow"] = ProjectileAsset{
		Image:          ScaleImg(ebiten.NewImageFromImage(im), 0.5),
		currentFrame:   frameCount,
		animationSpeed: 0,
	}
	return projectiles, nil
}

func loadTileset(data []byte, tileSize int, scale float64) (*Tileset, error) {
	im, err := loadImageFromBinaryPng(data)
	if err != nil {
		fmt.Println("Could not read assets!", err.Error())
		return nil, err
	}
	ebitenImage := ebiten.NewImageFromImage(im)
	return NewTileset(ebitenImage, int(im.Bounds().Dx()/tileSize), tileSize, scale)
}

func loadImageFromBinaryPng(dat []byte) (image.Image, error) {
	im, _, err := image.Decode(bytes.NewReader(dat))
	if err != nil {
		return nil, err
	}
	return im, nil
}

func readPngAsset(path string) (image.Image, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadImageFromBinaryPng(dat)
}
