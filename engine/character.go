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
}
type CharacterAsset struct {
	Animations     map[string]GameAssetAnimation
	Tileset        Tileset
	animationSpeed int
}

func (a *CharacterAsset) GetTile(activeAnimation string, frameCount int64) (*ebiten.Image, error) {
	animation, ok := a.Animations[activeAnimation]
	if !ok {
		return nil, fmt.Errorf("Unknown animation %s", activeAnimation)
	}

	animationFrame := int(frameCount/int64(a.animationSpeed)) % animation.FrameCount
	tileIdx := animation.StartTile + animationFrame
	subIm, err := a.Tileset.GetTile(tileIdx)
	if err != nil {
		return nil, err
	}
	return subIm, nil
}
