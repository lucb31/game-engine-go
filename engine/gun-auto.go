package engine

import (
	"fmt"
)

type AutoAimGun struct {
	*BasicGun
}

func NewAutoAimGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*AutoAimGun, error) {
	base, err := NewBasicGun(em, owner, proj, opts)
	if err != nil {
		return nil, err
	}
	gun := &AutoAimGun{BasicGun: base}
	return gun, nil
}

func (g *AutoAimGun) Shoot() error {
	if g.IsReloading() {
		return fmt.Errorf("Still Reloading...")
	}

	// Select target
	target := g.chooseTarget()
	if target == nil {
		return nil
	}

	// Spawn projectile
	proj, err := NewProjectile(g, g.em, g.projectileAsset)
	if err != nil {
		return err
	}
	proj.target = target
	g.em.AddEntity(proj)
	g.lastProjectileFired = g.em.GetIngameTime()
	return nil
}
