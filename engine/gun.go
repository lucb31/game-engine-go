package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type Gun interface {
	Shoot() error
	FireRange() float64
	Power() float64
	IsReloading() bool
	Owner() GameEntity
}

type BasicGunOpts struct {
	FireRatePerSecond float64
	FireRange         float64
	Damage            float64
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
	damage            float64
}

// Determine power of owner entity. If not available use gun damage
func (g *BasicGun) Power() float64 {
	if atk, isAttacker := g.owner.(damage.Attacker); isAttacker {
		return atk.Power()
	}
	return g.damage
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

func (g *BasicGun) chooseTarget() GameEntity {
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

func NewBasicGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*BasicGun, error) {
	gun := &BasicGun{em: em, owner: owner, projectileAsset: proj}
	// Init defaults
	gun.damage = 30
	gun.fireRatePerSecond = 1.5
	gun.fireRange = 100

	// Parse opts
	if opts.FireRatePerSecond > 0 {
		gun.fireRatePerSecond = opts.FireRatePerSecond
	}
	if opts.FireRange > 0 {
		gun.fireRange = opts.FireRange
	}
	if opts.Damage > 0 {
		gun.damage = opts.Damage
	}
	return gun, nil
}
