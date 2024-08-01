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
	world               *GameWorld
	orientation         Orientation
	shape               *cp.Shape
	asset               *CharacterAsset
	projectileAsset     *ProjectileAsset
	lastProjectileFired time.Time
	animation           string
}

const (
	playerVelocity          = 50
	playerFireRatePerSecond = float64(1.3)
)

func NewPlayer(world *GameWorld, asset *CharacterAsset, projectileAsset *ProjectileAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, orientation: South, projectileAsset: projectileAsset}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 10, Y: 10})
	playerBody.UserData = p
	playerBody.SetVelocityUpdateFunc(p.calculateVelocity)
	p.shape = cp.NewBox(playerBody, 16, 16, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(PlayerCollision))

	return p, nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	p.asset.Draw(screen, p.animation, p.shape.Body().Position())
}

func (p *Player) shoot() {
	now := time.Now()
	duration := float64(time.Second) / playerFireRatePerSecond
	if now.Sub(p.lastProjectileFired) < time.Duration(duration) {
		fmt.Println("Reloading...")
		return
	}
	fmt.Println("Shooting!")
	projectilePos := p.shape.Body().Position()
	offset := 25.0
	switch p.orientation {
	case East:
		projectilePos.Add(cp.Vector{offset, 0})
	case West:
		projectilePos.Add(cp.Vector{-offset, 0})
	case North:
		projectilePos.Add(cp.Vector{0, -offset})
	case South:
		projectilePos.Add(cp.Vector{0, offset})
	}
	proj, err := NewProjectile(p.world, p.projectileAsset, projectilePos, p.orientation)
	if err != nil {
		fmt.Println("Could not shoot projectile")
		return
	}
	p.world.addObject(proj)
	p.lastProjectileFired = time.Now()
}

func (p *Player) Destroy() {
	fmt.Println("ERROR: Cannot destroy player")
}

func (p *Player) Id() GameEntityId      { return p.id }
func (p *Player) SetId(id GameEntityId) { p.id = id }
func (p *Player) Shape() *cp.Shape      { return p.shape }

func (p *Player) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.shoot()
	}
	// Smoothen velocity
	velocity := body.Velocity()
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		velocity.Y = max(-playerVelocity, velocity.Y-playerVelocity*0.1)
		p.orientation = North
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		velocity.Y = min(playerVelocity, velocity.Y+playerVelocity*0.1)
		p.orientation = South
	} else {
		velocity.Y = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		velocity.X = -playerVelocity
		p.orientation = West
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
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
