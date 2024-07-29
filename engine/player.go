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
	world       *GameWorld
	orientation Orientation
	Shape       *cp.Shape
	asset       *CharacterAsset
}

const (
	playerVelocity = 50
	playerTileSize = 48
)

func NewPlayer(world *GameWorld, asset *CharacterAsset) (*Player, error) {
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 10, Y: 10})
	playerShape := cp.NewBox(playerBody, 16, 16, 0)
	playerShape.SetElasticity(0)
	playerShape.SetFriction(0)

	return &Player{world: world, asset: asset, Shape: playerShape, orientation: South}, nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	// Determine active animation based on current velocity & orientation
	activeAnimation := "idle_"
	if p.Shape.Body().Velocity().Length() > 0.1 {
		activeAnimation = "walk_"
	}
	activeAnimation += string(p.orientation)

	subIm, err := p.asset.GetTile(activeAnimation)
	if err != nil {
		fmt.Println("Error animating player", err.Error())
		return
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.Shape.Body().Position().X, p.Shape.Body().Position().Y)
	screen.DrawImage(subIm, &op)
}

func (p *Player) Update() {
	p.readMovementInputs()
}

func (p *Player) readMovementInputs() {
	// Smoothen velocity
	velocity := p.Shape.Body().Velocity()
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

	p.Shape.Body().SetVelocityVector(velocity)
}
