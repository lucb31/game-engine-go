package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type GameEntity interface {
	Draw(*ebiten.Image)
	Update()
}

type GameObj struct {
	asset           *CharacterAsset
	activeAnimation string
}

type GameAssetAnimation struct {
	StartTile  int
	FrameCount int
	Flip       bool
}
type CharacterAsset struct {
	Animations     map[string]GameAssetAnimation
	Tileset        Tileset
	animationSpeed int
	currentFrame   *int64
}

func (a *CharacterAsset) GetTile(activeAnimation string) (*ebiten.Image, error) {
	animation, ok := a.Animations[activeAnimation]
	if !ok {
		return nil, fmt.Errorf("Unknown animation %s", activeAnimation)
	}

	animationFrame := int(*a.currentFrame/int64(a.animationSpeed)) % animation.FrameCount
	tileIdx := animation.StartTile + animationFrame
	subIm, err := a.Tileset.GetTile(tileIdx)
	if err != nil {
		return nil, err
	}
	if animation.Flip {
		return flipHorizontal(subIm), nil
	}
	return subIm, nil
}

func flipHorizontal(source *ebiten.Image) *ebiten.Image {
	result := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(float64(source.Bounds().Dx()), 0)
	result.DrawImage(source, op)
	return result
}
