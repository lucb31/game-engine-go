package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type Orientation uint8

const (
	East Orientation = iota
	South
	West
	North
)

type Player struct {
	world       *GameWorld
	orientation Orientation
	Shape       *cp.Shape

	// Asset info
	Animations map[string]GameAssetAnimation
}

const (
	playerVelocity      = 50
	playerTileSize      = 48
	animationFrameCount = 6
	animationSpeed      = 6
	tilesPerRow         = 6
)

func NewPlayer(world *GameWorld) (*Player, error) {
	// Init player body & shape
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.SetPosition(cp.Vector{X: 10, Y: 10})
	playerShape := cp.NewBox(playerBody, 16, 16, 0)
	playerShape.SetElasticity(0)
	playerShape.SetFriction(0)
	// TODO: Dynamic velocity func
	// playerBody.SetVelocityUpdateFunc(calcPlayerVelocity)

	// TODO: Separate player asset from player logic
	animations := map[string]GameAssetAnimation{}
	animations["walk_horizontal"] = GameAssetAnimation{FrameCount: 6, TileIdx: 24}
	animations["walk_north"] = GameAssetAnimation{FrameCount: 6, TileIdx: 30}
	animations["walk_south"] = GameAssetAnimation{FrameCount: 6, TileIdx: 18}
	return &Player{world: world, Animations: animations, Shape: playerShape}, nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	// Animation: Determine tile idx
	var animationKey string
	flip := false
	switch p.orientation {
	case North:
		animationKey = "walk_north"
	case South:
		animationKey = "walk_south"
	case East:
		animationKey = "walk_horizontal"
	case West:
		animationKey = "walk_horizontal"
		flip = true
	}
	animation := p.Animations[animationKey]
	animationFrame := int(p.world.FrameCount/animationSpeed) % animation.FrameCount
	tileIdx := p.Animations[animationKey].TileIdx + animationFrame
	subIm, err := p.world.AssetManager.GetTile("player", tileIdx)
	if err != nil {
		fmt.Println("Error drawing player", err.Error())
		return
	}

	op := ebiten.DrawImageOptions{}
	// Offset size of player frame
	op.GeoM.Translate(-playerTileSize/2, -playerTileSize/2)
	op.GeoM.Translate(p.Shape.Body().Position().X, p.Shape.Body().Position().Y)
	if flip {
		screen.DrawImage(FlipHorizontal(subIm), &op)
	} else {
		screen.DrawImage(subIm, &op)
	}
}

func FlipHorizontal(source *ebiten.Image) *ebiten.Image {
	result := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(float64(source.Bounds().Dx()), 0)
	result.DrawImage(source, op)
	return result
}

func (p *Player) Update() {
	p.readMovementInputs()
}

//func calcPlayerVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
//body.Velocity()
//}

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
