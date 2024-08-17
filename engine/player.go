package engine

import (
	"fmt"
	"time"

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
	id                  GameEntityId
	world               GameEntityManager
	orientation         Orientation
	shape               *cp.Shape
	asset               *CharacterAsset
	projectileAsset     *ProjectileAsset
	lastProjectileFired time.Time
	animation           string
	gun                 Gun
}

const (
	playerVelocity = 100
	playerWidth    = 32
	playerHeight   = 32
)

func NewPlayer(world GameEntityManager, asset *CharacterAsset, projectileAsset *ProjectileAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, orientation: South, projectileAsset: projectileAsset}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 70, Y: 15})
	playerBody.UserData = p
	playerBody.SetVelocityUpdateFunc(p.calculateVelocity)

	// Collision model
	p.shape = cp.NewBox(playerBody, playerWidth, playerHeight, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(PlayerCollision))
	p.shape.SetFilter(PlayerCollisionFilter())

	var err error
	gunOpts := BasicGunOpts{FireRatePerSecond: 1.3}
	p.gun, err = NewSimpleGun(world, p, projectileAsset, &p.orientation, gunOpts)
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

func (p *Player) shoot() {
}

func (p *Player) Destroy() error {
	return fmt.Errorf("ERROR: Cannot destroy player")
}

func (p *Player) Id() GameEntityId      { return p.id }
func (p *Player) SetId(id GameEntityId) { p.id = id }
func (p *Player) Shape() *cp.Shape      { return p.shape }

func (p *Player) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !p.gun.IsReloading() {
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
