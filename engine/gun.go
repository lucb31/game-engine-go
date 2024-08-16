package engine

import "fmt"

type Gun interface {
	Shoot() error
	FireRange() float64
	IsReloading() bool
	Owner() GameEntity
}

type BasicGunOpts struct {
	FireRatePerSecond float64
	FireRange         float64
}

type BasicGun struct {
	em              GameEntityManager
	owner           GameEntity
	projectileAsset *ProjectileAsset

	// Runtime
	lastProjectileFired float64

	// Opts
	fireRatePerSecond float64
	fireRange         float64
}

func (g *BasicGun) Owner() GameEntity  { return g.owner }
func (g *BasicGun) FireRange() float64 { return g.fireRange }

func (g *BasicGun) IsReloading() bool {
	now := g.em.GetIngameTime()
	diff := now - g.lastProjectileFired
	if diff < 1/g.fireRatePerSecond {
		return true
	}
	return false
}

func NewBasicGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*BasicGun, error) {
	gun := &BasicGun{em: em, owner: owner, projectileAsset: proj}
	// Parse opts
	if opts.FireRatePerSecond > 0 {
		gun.fireRatePerSecond = opts.FireRatePerSecond
	} else {
		gun.fireRatePerSecond = 1.5
	}
	if opts.FireRange > 0 {
		gun.fireRange = opts.FireRange
	} else {
		gun.fireRange = 100
	}
	return gun, nil
}

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
	proj, err := NewProjectileWithOrientation(g, g.em, g.projectileAsset, *g.orientation)
	if err != nil {
		return err
	}
	g.em.AddEntity(proj)
	g.lastProjectileFired = g.em.GetIngameTime()
	return nil
}
