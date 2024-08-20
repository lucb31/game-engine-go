package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
)

type Orientation string

const (
	East  Orientation = "east"
	South             = "south"
	West              = "west"
	North             = "north"
)

type Player struct {
	id              GameEntityId
	world           GameEntityManager
	orientation     Orientation
	shape           *cp.Shape
	asset           *CharacterAsset
	projectileAsset *ProjectileAsset
	animation       string

	// ATK
	gun Gun

	controller PlayerController

	GameEntityStats
}

type GameEntityStats struct {
	armor         float64
	health        float64
	maxHealth     float64
	movementSpeed float64
	power         float64
}

type GameEntityStatReader interface {
	Armor() float64
	Health() float64
	MaxHealth() float64
	Power() float64
	MovementSpeed() float64
}

type GameEntityStatWriter interface {
	SetArmor(v float64)
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
func (s *GameEntityStats) Health() float64        { return s.health }
func (s *GameEntityStats) MaxHealth() float64     { return s.maxHealth }
func (s *GameEntityStats) Power() float64         { return s.power }
func (s *GameEntityStats) MovementSpeed() float64 { return s.movementSpeed }

func (s *GameEntityStats) SetArmor(v float64)         { s.armor = v }
func (s *GameEntityStats) SetHealth(h float64)        { s.health = h }
func (s *GameEntityStats) SetPower(v float64)         { s.power = v }
func (s *GameEntityStats) SetMaxHealth(v float64)     { s.maxHealth = v }
func (s *GameEntityStats) SetMovementSpeed(v float64) { s.movementSpeed = v }

const (
	playerWidth  = 32
	playerHeight = 32
)

func NewPlayer(world GameEntityManager, asset *CharacterAsset, projectileAsset *ProjectileAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, orientation: East, projectileAsset: projectileAsset}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 1470, Y: 820})
	playerBody.UserData = p
	playerBody.SetVelocityUpdateFunc(p.calculateVelocity)
	p.orientation = East

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
	p.movementSpeed = 150
	p.maxHealth = 100
	p.power = 30
	p.health = p.maxHealth

	// Init gun
	var err error
	gunOpts := BasicGunOpts{FireRatePerSecond: 1.3, FireRange: 250.0}
	p.gun, err = NewAutoAimGun(world, p, projectileAsset, gunOpts)
	if err != nil {
		return nil, err
	}

	// Init input controller
	p.controller, err = NewKeyboardPlayerController()
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
	return p.asset.Draw(t, p.animation, p.shape)
}

func (p *Player) DrawPlayerStats(t RenderingTarget) error {
	ebitenutil.DebugPrintAt(t.Screen(), "Player stats", t.Screen().Bounds().Dx()-125, 20)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Power %.2f", p.Power()), t.Screen().Bounds().Dx()-125, 35)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Health %.2f", p.Health()), t.Screen().Bounds().Dx()-125, 50)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Max Health %.2f", p.MaxHealth()), t.Screen().Bounds().Dx()-125, 65)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Speed %.2f", p.MovementSpeed()), t.Screen().Bounds().Dx()-125, 80)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Armor %.2f", p.Armor()), t.Screen().Bounds().Dx()-125, 95)
	return nil
}

func (p *Player) Destroy() error {
	p.world.EndGame()
	return nil
}

func (p *Player) OnPlayerHit(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	_, b := arb.Bodies()
	npc, ok := b.UserData.(*NpcEntity)
	if !ok {
		fmt.Println("Collsion handler error: Expected npc but did not receive one")
		return true
	}
	_, err := p.world.DamageModel().ApplyDamage(npc, p, p.world.GetIngameTime())
	if err != nil {
		fmt.Println("Error during player npc collision damage calc", err.Error())
		return true
	}
	npc.Destroy()

	return false
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
	velocity := p.controller.CalcVelocity(body.Velocity(), p.MovementSpeed())
	body.SetVelocityVector(velocity)

	// Update orientation
	if velocity.Length() > 0.0 {
		p.orientation = calculateOrientation(velocity)
	}
	p.animation = calculateWalkingAnimation(velocity, p.orientation)
}
