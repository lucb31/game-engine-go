package engine

import (
	"fmt"
)

type AutoAimGun struct {
	*BasicGun
	*GunTargetController
}

func NewAutoAimGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*AutoAimGun, error) {
	// Init base
	base, err := newBasicGun(em, owner, proj, opts)
	if err != nil {
		return nil, err
	}
	gun := &AutoAimGun{BasicGun: base}

	// Init target controller
	if gun.GunTargetController, err = newGunTargetController(gun); err != nil {
		return nil, err
	}

	return gun, nil
}

func (g *AutoAimGun) Shoot() error {
	if g.IsReloading() {
		return fmt.Errorf("Still Reloading...")
	}

	// Select targets
	targets := g.chooseTargets(g.projectileCount)
	if len(targets) == 0 {
		return nil
	}
	// Spawn projectile for every target
	for _, target := range targets {
		proj, err := NewProjectile(g, g.projectileAsset)
		if err != nil {
			return err
		}
		proj.SetTarget(target)
		g.em.AddEntity(proj)
	}
	// Set one common reload timer
	g.reloadTimeout.Set(1 / g.FireRate())

	if err := g.PlayShootSE(); err != nil {
		return err
	}
	// Play animation once
	if g.playShootAnimation != nil {
		// Calculate projectile orientation relative to owner
		// FIX: Currently simply using first target
		direction := targets[0].Body().Position().Sub(g.owner.Shape().Body().Position())
		orientation := ^West
		if direction.X > 0 {
			orientation = West
		}
		g.playShootAnimation(g.FireRate(), orientation)
	}
	return nil
}
