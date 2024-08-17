package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
)

const DEBUG_RENDER_COLLISION_BOXES = true

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

func (a *CharacterAsset) Draw(screen *ebiten.Image, activeAnimation string, shape *cp.Shape) error {
	if DEBUG_RENDER_COLLISION_BOXES {
		a.DrawRectBoundingBox(screen, shape)
	}
	subIm, err := a.GetTile(activeAnimation)
	if err != nil {
		return fmt.Errorf("Error animating player: %s", err.Error())
	}
	op := ebiten.DrawImageOptions{}
	// Offset to make sure asset is drawn centered on current position
	op.GeoM.Translate(a.offsetX, a.offsetY)
	op.GeoM.Translate(shape.Body().Position().X, shape.Body().Position().Y)
	screen.DrawImage(subIm, &op)
	return nil
}

func (a *CharacterAsset) DrawOnCamera(cam Camera, activeAnimation string, shape *cp.Shape) error {
	if DEBUG_RENDER_COLLISION_BOXES {
		a.DrawRectBoundingBoxOnCamera(cam, shape)
	}
	subIm, err := a.GetTile(activeAnimation)
	if err != nil {
		return fmt.Errorf("Error animating player: %s", err.Error())
	}
	op := ebiten.DrawImageOptions{}
	// Offset to make sure asset is drawn centered on current position
	op.GeoM.Translate(a.offsetX, a.offsetY)
	op.GeoM.Translate(shape.Body().Position().X, shape.Body().Position().Y)
	cam.DrawImage(subIm, &op)
	return nil
}

// Draws Rect bounding box around shape position
func (a *CharacterAsset) DrawRectBoundingBox(screen *ebiten.Image, shape *cp.Shape) error {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	vector.StrokeRect(screen, float32(shape.Body().Position().X-width/2), float32(shape.Body().Position().Y-height/2), float32(width), float32(height), 2.5, color.RGBA{255, 0, 0, 255}, false)
	return nil
}

// Draws Rect bounding box around shape position
func (a *CharacterAsset) DrawRectBoundingBoxOnCamera(cam Camera, shape *cp.Shape) error {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	cam.StrokeRect(shape.Body().Position().X-width/2, shape.Body().Position().Y-height/2, float32(width), float32(height), 2.5, color.RGBA{255, 0, 0, 255}, false)
	return nil
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
