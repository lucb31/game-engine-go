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
}

const (
	playerVelocity          = 100
	playerWidth             = 32
	playerHeight            = 32
	playerFireRatePerSecond = float64(1.3)
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
	return p, nil
}

func PlayerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, PlayerCategory, PlayerCategory|NpcCategory|OuterWallsCategory|TowerCategory)
}

func (p *Player) Draw(screen *ebiten.Image) {
	p.asset.DrawRectBoundingBox(screen, p.shape)
	p.asset.Draw(screen, p.animation, p.shape)
}

func (p *Player) shoot() {
	now := time.Now()
	duration := float64(time.Second) / playerFireRatePerSecond
	if now.Sub(p.lastProjectileFired) < time.Duration(duration) {
		// Still reloading
		return
	}
	// Spawn projectile at player position
	proj, err := NewProjectileWithOrientation(p, p.world, p.projectileAsset, p.orientation)
	if err != nil {
		fmt.Println("Could not shoot projectile")
		return
	}
	p.world.AddEntity(proj)
	p.lastProjectileFired = time.Now()
}

func (p *Player) Destroy() error {
	return fmt.Errorf("ERROR: Cannot destroy player")
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
