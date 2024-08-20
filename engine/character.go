package engine

import (
	"fmt"
	"image/color"

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
		return FlipHorizontal(subIm), nil
	}
	return subIm, nil
}

func (a *CharacterAsset) Draw(t RenderingTarget, activeAnimation string, shape *cp.Shape) error {
	if DEBUG_RENDER_COLLISION_BOXES {
		a.DrawRectBoundingBox(t, shape)
	}
	subIm, err := a.GetTile(activeAnimation)
	if err != nil {
		return fmt.Errorf("Error animating character: %s", err.Error())
	}
	op := ebiten.DrawImageOptions{}
	// Offset to make sure asset is drawn centered on current position
	op.GeoM.Translate(a.offsetX, a.offsetY)
	op.GeoM.Translate(shape.Body().Position().X, shape.Body().Position().Y)
	t.DrawImage(subIm, &op)
	return nil
}

// Draws Rect bounding box around shape position
func (a *CharacterAsset) DrawRectBoundingBox(t RenderingTarget, shape *cp.Shape) error {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	t.StrokeRect(shape.Body().Position().X-width/2, shape.Body().Position().Y-height/2, float32(width), float32(height), 2.5, color.RGBA{255, 0, 0, 255}, false)
	return nil
}

func (a *CharacterAsset) DrawHealthbar(t RenderingTarget, shape *cp.Shape, health, maxHealth float64) {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	// Outline
	t.StrokeRect(
		shape.Body().Position().X-width/2,
		shape.Body().Position().Y-height/2-12,
		float32(width),
		6,
		1,
		color.RGBA{255, 255, 255, 255},
		false,
	)
	// FILL
	maxWidth := width - 4
	filledWidth := float32(health / maxHealth * maxWidth)
	t.StrokeRect(
		shape.Body().Position().X-width/2+1,
		shape.Body().Position().Y-height/2-10,
		filledWidth,
		2,
		2,
		color.RGBA{255, 0, 0, 255},
		false,
	)
}

func calculateWalkingAnimation(vel cp.Vector, orientation Orientation) string {
	animation := "idle_"
	if vel.Length() > 5.0 {
		animation = "walk_"
	}

	// Append horizontal orientation
	orientationString := "east"
	if orientation&West == 0 {
		orientationString = "west"
	}
	return animation + orientationString
}

func calculateOrientation(vel cp.Vector) Orientation {
	res := Orientation(uint(0))
	if vel.X > 0 {
		res = res | West
	}
	if vel.Y > 0 {
		res = res | North
	}
	return res
}
