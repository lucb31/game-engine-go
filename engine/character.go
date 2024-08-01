package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

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
	offsetX        float64
	offsetY        float64
}

func (a *CharacterAsset) GetTile(activeAnimation string) (*ebiten.Image, error) {
	// Get animation tile
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

func (a *CharacterAsset) Draw(screen *ebiten.Image, activeAnimation string, position cp.Vector) error {
	subIm, err := a.GetTile(activeAnimation)
	if err != nil {
		return fmt.Errorf("Error animating player", err.Error())
	}
	op := ebiten.DrawImageOptions{}
	// Offset to make sure asset is drawn centered on current position
	op.GeoM.Translate(a.offsetX, a.offsetY)
	op.GeoM.Translate(position.X, position.Y)
	screen.DrawImage(subIm, &op)
	return nil
}

func flipHorizontal(source *ebiten.Image) *ebiten.Image {
	result := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(float64(source.Bounds().Dx()), 0)
	result.DrawImage(source, op)
	return result
}

func calculateWalkingAnimation(vel cp.Vector, orientation Orientation) string {
	animation := "idle_"
	if vel.Length() > 5.0 {
		animation = "walk_"
	}
	return animation + string(orientation)
}

func calculateOrientation(vel cp.Vector) Orientation {
	if vel.Y > 5 {
		return South
	} else if vel.Y < -5 {
		return North
	}
	if vel.X > 5 {
		return East
	} else if vel.X < -5 {
		return West
	}
	return East
}
