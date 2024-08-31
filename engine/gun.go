package engine

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type ShootingAnimationCallback func(float64, Orientation)
type Gun interface {
	PositionProvider
	Shoot() error
	FireRange() float64
	// Nr of projectiles per second
	FireRate() float64
	ProjectileCount() int
	Power() float64
	IsReloading() bool
	Owner() GameEntity

	SetShootingAnimationCallback(ShootingAnimationCallback)
	SetProjectileCount(int)
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
	reloadTimeout Timeout

	// Opts
	fireRatePerSecond float64
	fireRange         float64
	damage            float64
	projectileCount   int

	// Callback to play shooting animation
	playShootAnimation ShootingAnimationCallback
}

func newBasicGun(em GameEntityManager, owner GameEntity, proj *ProjectileAsset, opts BasicGunOpts) (*BasicGun, error) {
	gun := &BasicGun{em: em, owner: owner, projectileAsset: proj}
	var err error
	if gun.reloadTimeout, err = NewIngameTimeout(em); err != nil {
		return nil, err
	}

	// Init defaults
	gun.damage = 30
	gun.fireRatePerSecond = 1.5
	gun.fireRange = 100
	gun.projectileCount = 1

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
func (g *BasicGun) Owner() GameEntity    { return g.owner }
func (g *BasicGun) Position() cp.Vector  { return g.owner.Shape().Body().Position() }
func (g *BasicGun) FireRange() float64   { return g.fireRange }
func (g *BasicGun) ProjectileCount() int { return g.projectileCount }
func (g *BasicGun) IsReloading() bool {
	return !g.reloadTimeout.Done()
}

func (g *BasicGun) SetShootingAnimationCallback(playShootAnimation ShootingAnimationCallback) {
	g.playShootAnimation = playShootAnimation
}
func (g *BasicGun) SetProjectileCount(count int) { g.projectileCount = count }

func (g *BasicGun) PlayShootSE() error {
	reader := bytes.NewReader(assets.ShootOGG)
	stream, err := vorbis.DecodeWithoutResampling(reader)
	if err != nil {
		return err
	}
	player, err := audio.CurrentContext().NewPlayer(stream)
	if err != nil {
		return err
	}
	player.SetVolume(SFX_VOLUME)
	player.Play()
	return nil
}
