package engine

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
)

type Orientation uint8

const (
	West Orientation = 1 << iota
	North
)

type Player struct {
	// Dependencies
	id               GameEntityId
	world            GameEntityManager
	controller       PlayerController
	animationManager AnimationController
	asset            *CharacterAsset
	projectileAsset  *ProjectileAsset

	// Physics
	shape *cp.Shape

	// Damage model
	gun Gun
	GameEntityStats

	// Eyeframes
	lastHitAt float64
}

type GameEntityStats struct {
	armor         float64
	atkSpeed      float64
	health        float64
	maxHealth     float64
	movementSpeed float64
	power         float64
}

func DefaultGameEntityStats() GameEntityStats {
	return GameEntityStats{0, 1.0, 100.0, 100.0, 100.0, 30.0}
}

type GameEntityStatReader interface {
	Armor() float64
	AtkSpeed() float64
	Health() float64
	MaxHealth() float64
	Power() float64
	MovementSpeed() float64
}

type GameEntityStatWriter interface {
	SetArmor(v float64)
	SetAtkSpeed(v float64)
	SetHealth(h float64)
	SetPower(v float64)
	SetMaxHealth(v float64)
	SetMovementSpeed(v float64)
}

type GameEntityStatReadWriter interface {
	GameEntityStatReader
	GameEntityStatWriter
}

func (s *GameEntityStats) Armor() float64         { return s.armor }
func (s *GameEntityStats) AtkSpeed() float64      { return s.atkSpeed }
func (s *GameEntityStats) Health() float64        { return s.health }
func (s *GameEntityStats) MaxHealth() float64     { return s.maxHealth }
func (s *GameEntityStats) Power() float64         { return s.power }
func (s *GameEntityStats) MovementSpeed() float64 { return s.movementSpeed }

func (s *GameEntityStats) SetArmor(v float64)         { s.armor = v }
func (s *GameEntityStats) SetAtkSpeed(v float64)      { s.atkSpeed = v }
func (s *GameEntityStats) SetHealth(h float64)        { s.health = math.Min(h, s.maxHealth) }
func (s *GameEntityStats) SetPower(v float64)         { s.power = v }
func (s *GameEntityStats) SetMaxHealth(v float64)     { s.maxHealth = v }
func (s *GameEntityStats) SetMovementSpeed(v float64) { s.movementSpeed = v }

const (
	playerWidth                    = 40
	playerHeight                   = 40
	invulnerableForSecondsAfterHit = 0.5
)

func NewPlayer(world GameEntityManager, asset *CharacterAsset, projectileAsset *ProjectileAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, projectileAsset: projectileAsset}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 1470, Y: 820})
	playerBody.UserData = p
	playerBody.SetVelocityUpdateFunc(p.calculateVelocity)

	// Collision model
	p.shape = cp.NewBox(playerBody, playerWidth, playerHeight, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(PlayerCollision))
	p.shape.SetFilter(PlayerCollisionFilter())

	// Register npc collision handler
	ch := world.Space().NewCollisionHandler(cp.CollisionType(PlayerCollision), cp.CollisionType(NpcCollision))
	ch.BeginFunc = p.OnPlayerHit

	// Init stats
	p.GameEntityStats = DefaultGameEntityStats()
	p.movementSpeed = 150
	p.atkSpeed = 1.0

	// Init gun
	var err error
	gunOpts := BasicGunOpts{FireRange: 250.0}
	p.gun, err = NewAutoAimGun(world, p, projectileAsset, gunOpts)
	if err != nil {
		return nil, err
	}
	// Play shooting animation when gun shoots
	p.gun.SetShootingAnimationCallback(func(f float64, orientation Orientation) {
		p.animationManager.Play("shoot", 2, orientation)
	})

	// Init animation controller
	p.animationManager, err = NewAnimationManager(p.asset)
	if err != nil {
		return nil, err
	}

	// Init input controller
	p.controller, err = NewKeyboardPlayerController(p.animationManager)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func PlayerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, PlayerCategory, PlayerCategory|NpcCategory|OuterWallsCategory|TowerCategory)
}

func (p *Player) Draw(t RenderingTarget) error {
	p.asset.DrawHealthbar(t, p.shape, p.health, p.maxHealth)
	if err := p.DrawPlayerStats(t); err != nil {
		return err
	}
	// Play death animation loop when dead
	if p.health <= 0 {
		err := p.animationManager.Loop("dead", p.controller.Orientation())
		if err != nil {
			fmt.Println("could not loop death animation", err.Error())
		}
	}
	return p.animationManager.Draw(t, p.shape)
}

func (p *Player) DrawPlayerStats(t RenderingTarget) error {
	ebitenutil.DebugPrintAt(t.Screen(), "Player stats", t.Screen().Bounds().Dx()-125, 20)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Power %.2f", p.Power()), t.Screen().Bounds().Dx()-125, 35)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Health %.2f", p.Health()), t.Screen().Bounds().Dx()-125, 50)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Max Health %.2f", p.MaxHealth()), t.Screen().Bounds().Dx()-125, 65)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Speed %.2f", p.MovementSpeed()), t.Screen().Bounds().Dx()-125, 80)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Armor %.2f", p.Armor()), t.Screen().Bounds().Dx()-125, 95)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("AtkSpeed %.2f", p.AtkSpeed()), t.Screen().Bounds().Dx()-125, 110)
	return nil
}

func (p *Player) Destroy() error {
	// Play dying animation
	err := p.animationManager.Play("die", 5, p.controller.Orientation())
	if err != nil {
		fmt.Println("Could not play dying animation", err.Error())
	}

	// Trigger game over
	p.world.EndGame()
	return nil
}

func (p *Player) OnPlayerHit(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	_, b := arb.Bodies()
	npc, ok := b.UserData.(*NpcEntity)
	if !ok {
		fmt.Println("Collsion handler error: Expected npc but did not receive one")
		return false
	}
	_, err := p.world.DamageModel().ApplyDamage(npc, p, p.world.GetIngameTime())
	if err != nil {
		fmt.Println("Error during player npc collision damage calc", err.Error())
		return false
	}

	// Play on hit animation
	err = p.animationManager.Play("hit", 5, p.controller.Orientation())
	if err != nil {
		fmt.Println("Could not play on hit animation", err.Error())
	}

	// Register eyeframe timeout
	p.lastHitAt = p.world.GetIngameTime()

	npc.Destroy()
	return false
}

// Player invulnerable for a brief period after being hit
func (p *Player) IsVulnerable() bool {
	if p.lastHitAt == 0 {
		return true
	}
	diff := p.world.GetIngameTime() - p.lastHitAt
	return diff > invulnerableForSecondsAfterHit
}

func (p *Player) Id() GameEntityId      { return p.id }
func (p *Player) SetId(id GameEntityId) { p.id = id }
func (p *Player) Shape() *cp.Shape      { return p.shape }
func (p *Player) LootTable() *LootTable { return EmptyLootTable() }

func (p *Player) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Automatically shoot
	if !p.gun.IsReloading() {
		if err := p.gun.Shoot(); err != nil {
			fmt.Println("Error when trying to shoot player gun", err.Error())
		}
	}
	velocity := p.controller.CalcVelocity(p.MovementSpeed(), p.world.GetIngameTime())
	body.SetVelocityVector(velocity)

}
