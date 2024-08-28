package engine

import (
	"log"
	"math"

	"github.com/jakecoffman/cp"
)

type PositionProvider interface {
	Position() cp.Vector
}

type FollowingCamera struct {
	BaseCamera
	target PositionProvider

	// In locked mode the camera will just copy the targets position
	// In unlocked mode the camera will move towards the target position
	locked bool
}

func NewFollowingCamera(width, height int) (*FollowingCamera, error) {
	base, err := NewBaseCamera(width, height)
	if err != nil {
		return nil, err
	}
	cam := &FollowingCamera{BaseCamera: *base}
	cam.Body().SetVelocityUpdateFunc(cam.calcVelocity)
	return cam, nil
}

func (c *FollowingCamera) calcVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	if c.target == nil {
		body.SetVelocity(0, 0)
		return
	}
	targetPos := c.target.Position()
	// Snap on target if below threshold
	// NOTE: Make sure distance is scaled with timestep
	if c.locked || body.Position().Near(targetPos, 650.0*dt) {
		body.SetPosition(targetPos)
		body.SetVelocity(0, 0)
		c.locked = true
		return
	}

	maxSpeed := 500.0
	distance := targetPos.Sub(body.Position())
	direction := distance.Normalize()
	vel := direction.Mult(maxSpeed)
	body.SetVelocityVector(vel)
}

func (c *FollowingCamera) SetTarget(target PositionProvider) {
	c.target = target
	// Reset camera lock
	c.locked = false
}

// Backup: If we want to smoothen camera movement
func easedSpeed(distance cp.Vector, maxSpeed float64) {
	// Calculate eased speed
	maxDistanceSq := 10000.0
	fractionOfMaxDistance := math.Min(1, distance.LengthSq()/maxDistanceSq)
	speed := EaseInOutCubic(fractionOfMaxDistance) * maxSpeed
	log.Println(distance.LengthSq(), fractionOfMaxDistance, speed)
}
