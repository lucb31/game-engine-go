package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type ShootingAnimationCallback func(float64, Orientation)
type Gun interface {
	Shoot() error
	FireRange() float64
	// Nr of projectiles per second
	FireRate() float64
	Power() float64
	IsReloading() bool
	// Returns game tick of next round to be shot. Required to sync shooting animation
	RemainingReloadTime() float64
	Owner() GameEntity
	SetShootingAnimationCallback(ShootingAnimationCallback)
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

	// Callback to play shooting animation
	playShootAnimation ShootingAnimationCallback
}

// Determine power of owner entity. If not available use gun damage
func (g *BasicGun) Power() float64 {
	if atk, isAttacker := g.owner.(damage.Attacker); isAttacker {
		return atk.Power()
	}
	return g.damage
}

// Determine by atk speed of owner
func (g *BasicGun) FireRate() float64 {
	if atk, isAttacker := g.owner.(damage.Attacker); isAttacker {
		return atk.AtkSpeed()
	}
	return g.fireRatePerSecond
}

func (g *BasicGun) Owner() GameEntity  { return g.owner }
func (g *BasicGun) FireRange() float64 { return g.fireRange }

func (g *BasicGun) RemainingReloadTime() float64 {
	nextBulletAt := g.lastProjectileFired + 1/g.FireRate()
	now := g.em.GetIngameTime()
	return nextBulletAt - now
}

func (g *BasicGun) IsReloading() bool {
	return g.RemainingReloadTime() > 0
}

func (g *BasicGun) SetShootingAnimationCallback(playShootAnimation ShootingAnimationCallback) {
	g.playShootAnimation = playShootAnimation
}

func (g *BasicGun) chooseTarget() GameEntity {
	query := g.owner.Shape().Space().PointQueryNearest(g.owner.Shape().Body().Position(), g.fireRange, gunTargetCollisionFilter)
	if query.Shape == nil {
		return nil
	}
	npc, ok := query.Shape.Body().UserData.(*NpcEntity)
	if !ok {
		fmt.Println("Expected npc target, but found something else", query.Shape.Body().UserData)
		return nil
	}
	return npc
}

var gunTargetCollisionFilter = cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, NpcCategory)

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
