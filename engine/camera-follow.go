package engine

import (
	"github.com/jakecoffman/cp"
)

type FollowingCamera struct {
	BaseCamera
	target GameEntity
}

func NewFollowingCamera(width, height int, target GameEntity) (*FollowingCamera, error) {
	base, err := NewBaseCamera(width, height)
	if err != nil {
		return nil, err
	}
	cam := &FollowingCamera{BaseCamera: *base, target: target}
	cam.Body().SetPositionUpdateFunc(cam.calcPosition)
	return cam, nil
}

func (c *FollowingCamera) calcPosition(body *cp.Body, dt float64) {
	body.SetPosition(c.target.Shape().Body().Position())
}
