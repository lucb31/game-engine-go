package engine

import (
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

const followCamMaxSpeed = 500.0

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
	// NOTE: Make sure near distance is scaled with timestep
	nearDistanceSq := (followCamMaxSpeed * dt) * (followCamMaxSpeed * dt)
	// Vector from camera center to targetPos
	distance := c.target.Position().Sub(c.CenterWorldPos())

	// Snap on target if below threshold
	if c.locked || distance.LengthSq() < nearDistanceSq {
		// Apply distance vector to current pos
		body.SetPosition(body.Position().Add(distance))
		body.SetVelocity(0, 0)
		c.locked = true
		return
	}

	direction := distance.Normalize()
	vel := direction.Mult(followCamMaxSpeed)
	body.SetVelocityVector(vel)
}

func (c *FollowingCamera) SetTarget(target PositionProvider) {
	c.target = target
	// Reset camera lock
	c.locked = false
}

// Backup: If we want to smoothen camera movement
func easedSpeed(distance cp.Vector, maxSpeed float64) float64 {
	// Calculate eased speed
	maxDistanceSq := 10000.0
	fractionOfMaxDistance := math.Min(1, distance.LengthSq()/maxDistanceSq)
	speed := EaseInOutCubic(fractionOfMaxDistance) * maxSpeed
	// log.Println(distance.LengthSq(), fractionOfMaxDistance, speed)
	return speed
}
