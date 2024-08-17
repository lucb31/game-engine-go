package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type FreeMovementCamera struct {
	BaseCamera
}

func NewFreeMovementCamera(width, height int) (*FreeMovementCamera, error) {
	base, err := NewBaseCamera(width, height)
	if err != nil {
		return nil, err
	}
	cam := &FreeMovementCamera{BaseCamera: *base}
	cam.Body().SetVelocityUpdateFunc(cam.calculateVelocity)
	return cam, nil
}

// Control camera with arrow keys
func (c *FreeMovementCamera) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	absoluteVel := 500.0
	velocity := body.Velocity()
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		velocity.Y = max(-absoluteVel, velocity.Y-absoluteVel*0.1)
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		velocity.Y = min(absoluteVel, velocity.Y+absoluteVel*0.1)
	} else {
		velocity.Y = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		velocity.X = -absoluteVel
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		velocity.X = absoluteVel
	} else {
		velocity.X = 0
	}
	// Update physics velocity
	body.SetVelocityVector(velocity)
}
