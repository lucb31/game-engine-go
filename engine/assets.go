package engine

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
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

func NewAssetManager(atp AnimationTimeProvider) (*AssetManagerImpl, error) {
	am := &AssetManagerImpl{}
	var err error

	am.Tilesets, err = loadEnvironmentTilesets()
	if err != nil {
		return nil, err
	}

	am.characterAssets, err = loadCharacterAssets(atp)
	if err != nil {
		return nil, err
	}

	am.projectileAssets, err = loadProjectileAssets(atp)
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
	asset, ok := a.characterAssets[identifier]
	if !ok {
		return nil, fmt.Errorf("Trying to access unknown asset %s", identifier)
	}
	// Initialize animation manager here.
	// NOTE: If we do this within the character asset, all assets will share the same animation controller
	// And animation would be synced
	animationManager, err := NewAnimationManager(&asset)
	if err != nil {
		return nil, err
	}
	asset.animationManager = animationManager
	var initialAnimation string
	for key := range asset.Animations {
		initialAnimation = key
		break
	}
	if err := asset.AnimationController().Loop(initialAnimation); err != nil {
		return nil, err
	}
	return &asset, nil
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
	Key                  string
	ImageData            []byte
	TileSizeX, TileSizeY int
}

// TODO: Load these from config file
var tileResources []tileResource = []tileResource{
	// Used for td
	{"plains", assets.Plains, 16, 16},
	{"darkdimension", assets.Darkdimension, 16, 16},
	{"props", assets.Props, 16, 16},
}

// Load tilesets for static resources
func loadEnvironmentTilesets() (map[string]Tileset, error) {
	tiles := map[string]Tileset{}
	for _, res := range tileResources {
		tileset, err := loadTileset(res.ImageData, res.TileSizeX, res.TileSizeY, 1.0)
		if err != nil {
			return nil, err
		}
		tiles[res.Key] = *tileset
	}
	return tiles, nil
}

type characterResource struct {
	Key                  string
	ImageData            []byte
	TileSizeX, TileSizeY int
	Animations           map[string]GameAssetAnimation
	OffsetX, OffsetY     float64
	Scale                float64
}

const defaultAnimationSpeed = 0.08

// TODO: Load these from config file
var characterResources []characterResource = []characterResource{
	// {
	// 	"player",
	// 	assets.Player,
	// 	48,
	// 	48,
	// 	map[string]GameAssetAnimation{
	// 		"walk": {24, 6, 0.2},
	// 		"idle": { 6, 6},
	// 		"dash": { 42, FrameCount: 4},
	// 	},
	// 	-44,
	// 	-60,
	// 	1.8,
	// },
	{
		"ranger",
		assets.Ranger,
		288,
		128,
		map[string]GameAssetAnimation{
			"walk":    {StartTile: 22, FrameCount: 10, Speed: defaultAnimationSpeed},
			"idle":    {StartTile: 0, FrameCount: 12, Speed: defaultAnimationSpeed},
			"dash":    {StartTile: 198, FrameCount: 11, Speed: defaultAnimationSpeed},
			"hit":     {StartTile: 330, FrameCount: 6, Speed: defaultAnimationSpeed},
			"die":     {StartTile: 352, FrameCount: 18, Speed: defaultAnimationSpeed},
			"dead":    {StartTile: 370, FrameCount: 1, Speed: defaultAnimationSpeed},
			"shoot":   {StartTile: 242, FrameCount: 14, Speed: 0.05},
			"harvest": {StartTile: 220, FrameCount: 10, Speed: defaultAnimationSpeed},
		},
		-115,
		-85,
		0.8,
	},
	{
		"npc-torch",
		assets.NpcTorch,
		192,
		192,
		map[string]GameAssetAnimation{
			"attack": {StartTile: 14, FrameCount: 6, Speed: defaultAnimationSpeed * 2},
			"idle":   {StartTile: 0, FrameCount: 7, Speed: defaultAnimationSpeed},
			"walk":   {StartTile: 7, FrameCount: 6, Speed: defaultAnimationSpeed},
		},
		-40,
		-37,
		0.4,
	},
	{
		"npc-orc",
		assets.Orc,
		100,
		100,
		map[string]GameAssetAnimation{
			"walk":   {StartTile: 8, FrameCount: 6, Speed: defaultAnimationSpeed},
			"idle":   {StartTile: 0, FrameCount: 6, Speed: defaultAnimationSpeed},
			"attack": {StartTile: 16, FrameCount: 6, Speed: defaultAnimationSpeed * 2},
		},
		-60,
		-60,
		1.2,
	},
	{
		"npc-slime",
		assets.Slime,
		24,
		24,
		map[string]GameAssetAnimation{
			"walk":   {StartTile: 0, FrameCount: 2, Speed: defaultAnimationSpeed},
			"idle":   {StartTile: 0, FrameCount: 2, Speed: defaultAnimationSpeed},
			"attack": {StartTile: 0, FrameCount: 2, Speed: defaultAnimationSpeed},
		},
		-16,
		-16,
		1.5,
	},
	{
		"tower-blue",
		assets.TowerBlue,
		256,
		192,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 4, Speed: defaultAnimationSpeed},
		},
		-28,
		-20,
		0.22,
	},
	{
		"tower-red",
		assets.TowerRed,
		256,
		192,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1, Speed: defaultAnimationSpeed},
		},
		-28,
		-30,
		0.22,
	},
	{
		"castle",
		assets.Castle,
		192,
		128,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1},
		},
		-96,
		-64,
		1.0,
	},
	{
		"tree_a",
		assets.TreeA,
		16,
		32,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1},
		},
		-16,
		-32,
		2.0,
	},
	{
		"tree_b",
		assets.TreeB,
		45,
		64,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1},
		},
		-22.5,
		-32,
		1.0,
	},
	{
		"tree_small",
		assets.TreeSmall,
		32,
		32,
		map[string]GameAssetAnimation{
			"idle": {StartTile: 0, FrameCount: 1},
		},
		-16,
		-16,
		1.0,
	},
	{
		"wood",
		assets.Wood,
		128,
		128,
		map[string]GameAssetAnimation{
			"idle":  {StartTile: 5, FrameCount: 1},
			"spawn": {StartTile: 0, FrameCount: 6, Speed: defaultAnimationSpeed},
		},
		-32,
		-40,
		0.5,
	},
}

// Load characters
func loadCharacterAssets(atp AnimationTimeProvider) (map[string]CharacterAsset, error) {
	characters := map[string]CharacterAsset{}
	for _, res := range characterResources {
		tileset, err := loadTileset(res.ImageData, res.TileSizeX, res.TileSizeY, res.Scale)
		if err != nil {
			return nil, err
		}
		asset, err := NewCharacterAsset(atp)
		if err != nil {
			return nil, err
		}
		asset.Tileset = *tileset
		asset.Animations = res.Animations
		asset.offsetX = res.OffsetX
		asset.offsetY = res.OffsetY
		characters[res.Key] = *asset
	}
	return characters, nil
}

func loadProjectileAssets(atp AnimationTimeProvider) (map[string]ProjectileAsset, error) {
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
		atp:            atp,
		animationSpeed: 2.0,
	}

	// Load arrow
	im, err = loadImageFromBinaryPng(assets.Arrow)
	if err != nil {
		return nil, err
	}
	projectiles["arrow"] = ProjectileAsset{
		Image:          ScaleImg(ebiten.NewImageFromImage(im), 0.3),
		animationSpeed: 0.0,
		atp:            atp,
	}
	return projectiles, nil
}

func loadTileset(data []byte, tileSizeX, tileSizeY int, scale float64) (*Tileset, error) {
	im, err := loadImageFromBinaryPng(data)
	if err != nil {
		log.Fatalln("Could not read assets!", err.Error())
		return nil, err
	}
	ebitenImage := ebiten.NewImageFromImage(im)
	return NewTileset(ebitenImage, tileSizeX, tileSizeY, scale)
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
