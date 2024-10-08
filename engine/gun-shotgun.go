package engine

import (
	"fmt"
	"math"

	"github.com/jakecoffman/cp"
)

type ShotGun struct {
	BasicGun
	*GunTargetController
	projectiles int
}

func NewShotGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*ShotGun, error) {
	base, err := newBasicGun(em, owner, proj, opts)
	if err != nil {
		return nil, err
	}
	// TODO: Opt for nr of projectiles
	gun := &ShotGun{BasicGun: *base, projectiles: 5}
	if gun.GunTargetController, err = newGunTargetController(gun); err != nil {
		return nil, err
	}
	return gun, nil
}

func (g *ShotGun) Shoot() error {
	if g.IsReloading() {
		return fmt.Errorf("Still Reloading...")
	}

	// Dont shoot if nothing in range
	target := g.chooseTarget()
	if target == nil {
		return nil
	}

	// Spawn projectiles
	for idx := range g.projectiles {
		angleInRad := math.Pi * float64(2*idx) / float64(g.projectiles)
		direction := cp.Vector{
			X: g.owner.Shape().Body().Position().X + math.Sin(angleInRad)*g.fireRange,
			Y: g.owner.Shape().Body().Position().Y + math.Cos(angleInRad)*g.fireRange,
		}
		proj, err := NewProjectile(g, g.projectileAsset)
		if err != nil {
			return err
		}
		proj.direction = direction
		g.em.AddEntity(proj)
	}
	g.reloadTimeout.Set(1 / g.FireRate())
	return nil
}
