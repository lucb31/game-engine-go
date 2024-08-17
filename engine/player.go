package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
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

	// DEF
	health float64
}

const (
	playerVelocity = 500
	playerWidth    = 32
	playerHeight   = 32
)

func NewPlayer(world GameEntityManager, asset *CharacterAsset, projectileAsset *ProjectileAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, orientation: South, projectileAsset: projectileAsset}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 530, Y: 402})
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

	// Game logic
	p.health = 40
	var err error
	gunOpts := BasicGunOpts{FireRatePerSecond: 1.3, FireRange: 250.0}
	p.gun, err = NewAutoAimGun(world, p, projectileAsset, gunOpts)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func PlayerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, PlayerCategory, PlayerCategory|NpcCategory|OuterWallsCategory|TowerCategory)
}

func (p *Player) Draw(t RenderingTarget) {
	p.asset.Draw(t, p.animation, p.shape)
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

func (p *Player) Id() GameEntityId         { return p.id }
func (p *Player) SetId(id GameEntityId)    { p.id = id }
func (p *Player) Shape() *cp.Shape         { return p.shape }
func (p *Player) Health() float64          { return p.health }
func (p *Player) Armor() float64           { return 0 }
func (p *Player) SetHealth(health float64) { p.health = health }

func (p *Player) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Automatically shoot
	if !p.gun.IsReloading() {
		if err := p.gun.Shoot(); err != nil {
			fmt.Println("Error when trying to shoot player gun", err.Error())
		}
	}
	// Smoothen velocity
	velocity := body.Velocity()
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		velocity.Y = max(-playerVelocity, velocity.Y-playerVelocity*0.1)
		p.orientation = North
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		velocity.Y = min(playerVelocity, velocity.Y+playerVelocity*0.1)
		p.orientation = South
	} else {
		velocity.Y = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		velocity.X = -playerVelocity
		p.orientation = West
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		velocity.X = playerVelocity
		p.orientation = East
	} else {
		velocity.X = 0
	}
	// Update physics velocity
	body.SetVelocityVector(velocity)
	// Update animation
	if velocity.Length() > 0.0 {
		p.orientation = calculateOrientation(velocity)
	}
	p.animation = calculateWalkingAnimation(velocity, p.orientation)
}
