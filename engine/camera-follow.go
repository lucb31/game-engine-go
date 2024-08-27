package engine

import (
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
	cam.Body().SetPositionUpdateFunc(cam.calcPosition)
	return cam, nil
}

func (c *FollowingCamera) calcPosition(body *cp.Body, dt float64) {
	if c.target == nil {
		return
	}
	body.SetPosition(c.target.Position())
}

func (c *FollowingCamera) SetTarget(target PositionProvider) {
	c.target = target
}
