package engine

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type GameAssetAnimation struct {
	StartTile  int
	FrameCount int
	Speed      float64
}

type CharacterAsset struct {
	animationManager AnimationController
	Animations       map[string]GameAssetAnimation
	Tileset          Tileset
	offsetX          float64
	offsetY          float64
	atp              AnimationTimeProvider
}

func NewCharacterAsset(atp AnimationTimeProvider) (*CharacterAsset, error) {
	a := &CharacterAsset{atp: atp}
	return a, nil
}

// Get animation tile
// Could also move this to animation manager
func (a *CharacterAsset) GetTile(animation *GameAssetAnimation, animationTile int, flip bool) (*ebiten.Image, error) {
	if animation == nil {
		return nil, fmt.Errorf("No animation provided")
	}
	tileIdx := animation.StartTile + animationTile
	subIm, err := a.Tileset.GetTile(tileIdx)
	if err != nil {
		return nil, err
	}
	if flip {
		return FlipHorizontal(subIm), nil
	}
	return subIm, nil
}

func (a *CharacterAsset) DrawAnimationTile(t RenderingTarget, shape *cp.Shape, animation *GameAssetAnimation, animationTile int, o Orientation) error {
	if DEBUG_RENDER_COLLISION_BOXES {
		DrawRectBoundingBox(t, shape.BB())
	}
	flip := o&West == 0
	subIm, err := a.GetTile(animation, animationTile, flip)
	if err != nil {
		return fmt.Errorf("Error animating character: %s", err.Error())
	}
	op := ebiten.DrawImageOptions{}
	// Offset to make sure asset is drawn centered on current position
	op.GeoM.Translate(a.offsetX, a.offsetY)
	op.GeoM.Translate(shape.Body().Position().X, shape.Body().Position().Y)
	// Opposed to map data we apply a linear filter to character assets
	// Without it there are flickering effects
	op.Filter = ebiten.FilterLinear
	t.DrawImage(subIm, &op)
	return nil
}

func (a *CharacterAsset) Draw(t RenderingTarget, shape *cp.Shape, o Orientation) error {
	return a.animationManager.Draw(t, shape, o)
}

func (a *CharacterAsset) Animation(animationKey string) (*GameAssetAnimation, error) {
	animation, ok := a.Animations[animationKey]
	if !ok {
		return nil, fmt.Errorf("Unknown animation %s", animationKey)
	}
	return &animation, nil
}

func (a *CharacterAsset) AnimationController() AnimationController { return a.animationManager }
func (a *CharacterAsset) AnimationTime() float64                   { return a.atp.AnimationTime() }

func (a *CharacterAsset) DrawHealthbar(t RenderingTarget, shape *cp.Shape, health, maxHealth float64) {
	// How many px the healthbar should take in height
	healthBarHeight := 6.0
	// How many px margin between character BB and bottom of healthbar
	healthBarMargin := 10.0

	// Calculate outline pos
	outlineTopLeft := cp.Vector{shape.BB().L, shape.BB().B - healthBarMargin - healthBarHeight}
	outlineBotRight := cp.Vector{shape.BB().R, shape.BB().B - healthBarMargin}

	// Draw fill
	width := shape.BB().R - shape.BB().L
	maxWidth := width - 4
	filledWidth := math.Max(0, health/maxHealth*maxWidth)
	fillBotRight := cp.Vector{outlineTopLeft.X + filledWidth, outlineBotRight.Y}
	t.StrokeRect(
		outlineTopLeft,
		fillBotRight,
		float32(healthBarHeight),
		color.RGBA{255, 0, 0, 255},
		false,
	)

	// strokeWidth := 2.0
	// TODO: Draw outline
	// Could not get this to work correctly with zoom. Rather move this to HUD instead of trying to fix now
	// t.StrokeRect(outlineTopLeft, outlineBotRight, float32(strokeWidth), color.RGBA{255, 255, 255, 255}, false)
}

func updateOrientation(prev Orientation, vel cp.Vector) Orientation {
	if math.Abs(vel.X) > 0 {
		if vel.X > 0 {
			prev = prev | West
		} else {
			prev = prev &^ West
		}
	}
	if math.Abs(vel.Y) > 0 {
		if vel.Y < 0 {
			prev = prev | North
		} else {
			prev = prev &^ North
		}
	}
	return prev
}
