package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type SimpleGun struct {
	BasicGun
	orientation *Orientation
}

func NewSimpleGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, orientation *Orientation, opts BasicGunOpts) (*SimpleGun, error) {
	base, err := NewBasicGun(em, owner, proj, opts)
	if err != nil {
		return nil, err
	}
	gun := &SimpleGun{BasicGun: *base, orientation: orientation}
	return gun, nil
}

func (g *SimpleGun) Shoot() error {
	if g.IsReloading() {
		return fmt.Errorf("Still Reloading...")
	}

	// Spawn projectile at owner position & orientation
	proj, err := NewProjectile(g, g.em, g.projectileAsset)
	if err != nil {
		return err
	}
	proj.direction = directionFromOrientationAndPos(*g.orientation, g.Owner().Shape().Body().Position())
	g.em.AddEntity(proj)
	g.lastProjectileFired = g.em.GetIngameTime()
	return nil
}

func directionFromOrientationAndPos(orientation Orientation, pos cp.Vector) cp.Vector {
	switch orientation {
	case North:
		return cp.Vector{pos.X, -1000}
	case South:
		return cp.Vector{pos.X, 1000}
	case East:
		return cp.Vector{1000, pos.Y}
	default:
		return cp.Vector{-1000, pos.Y}
	}
}
