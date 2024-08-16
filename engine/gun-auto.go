package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type AutoAimGun struct {
	BasicGun
}

func NewAutoAimGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*AutoAimGun, error) {
	base, err := NewBasicGun(em, owner, proj, opts)
	if err != nil {
		return nil, err
	}
	gun := &AutoAimGun{BasicGun: *base}
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

func (g *AutoAimGun) chooseTarget() GameEntity {
	query := g.owner.Shape().Space().PointQueryNearest(g.owner.Shape().Body().Position(), g.fireRange, cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, NpcCategory))
	if query.Shape == nil {
		return nil
	}
	npc, ok := query.Shape.Body().UserData.(*NpcEntity)
	if !ok {
		fmt.Println("Expected npc target, but found something else")
		return nil
	}
	return npc
}
