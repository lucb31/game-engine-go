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
	if body.Position().Near(targetPos, 10.0) {
		body.SetPosition(targetPos)
		body.SetVelocity(0, 0)
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
}

// Backup: If we want to smoothen camera movement
func easedSpeed(distance cp.Vector, maxSpeed float64) {
	// Calculate eased speed
	maxDistanceSq := 10000.0
	fractionOfMaxDistance := math.Min(1, distance.LengthSq()/maxDistanceSq)
	speed := EaseInOutCubic(fractionOfMaxDistance) * maxSpeed
	log.Println(distance.LengthSq(), fractionOfMaxDistance, speed)
}
