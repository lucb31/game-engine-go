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
	id          GameEntityId
	world       *GameWorld
	orientation Orientation
	shape       *cp.Shape
	asset       *CharacterAsset
}

const (
	playerVelocity = 50
	playerTileSize = 48
)

func NewPlayer(world *GameWorld, asset *CharacterAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, orientation: South}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 10, Y: 10})
	playerBody.UserData = p
	p.shape = cp.NewBox(playerBody, 16, 16, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(PlayerCollision))

	return p, nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	//fmt.Println("Drawing player at position", p.Shape.Body().Position())
	// Determine active animation based on current velocity & orientation
	activeAnimation := "idle_"
	if p.shape.Body().Velocity().Length() > 0.1 {
		activeAnimation = "walk_"
	}
	activeAnimation += string(p.orientation)

	subIm, err := p.asset.GetTile(activeAnimation)
	if err != nil {
		fmt.Println("Error animating player", err.Error())
		return
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.shape.Body().Position().X, p.shape.Body().Position().Y)
	screen.DrawImage(subIm, &op)
}

func (p *Player) Update() {
	p.readMovementInputs()
}

func (p *Player) Destroy() {
	fmt.Println("ERROR: Cannot destroy player")
}

func (p *Player) Id() GameEntityId { return p.id }
func (p *Player) Shape() *cp.Shape { return p.shape }

func (p *Player) readMovementInputs() {
	// Smoothen velocity
	velocity := p.shape.Body().Velocity()
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

	p.shape.Body().SetVelocityVector(velocity)
}
