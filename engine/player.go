package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Orientation uint8

const (
	East Orientation = iota
	South
	West
	North
)

type Player struct {
	posX        float64
	posY        float64
	world       *GameWorld
	velX        float64
	velY        float64
	orientation Orientation

	// Asset info
	Animations map[string]GameAssetAnimation
}

const (
	playerVelocity      = 1.5
	playerTileSize      = 48
	animationFrameCount = 6
	animationSpeed      = 6
	tilesPerRow         = 6
)

func NewPlayer(world *GameWorld) (*Player, error) {
	animations := map[string]GameAssetAnimation{}
	animations["walk_horizontal"] = GameAssetAnimation{FrameCount: 6, TileIdx: 24}
	animations["walk_north"] = GameAssetAnimation{FrameCount: 6, TileIdx: 30}
	animations["walk_south"] = GameAssetAnimation{FrameCount: 6, TileIdx: 18}
	return &Player{posX: 10, posY: 10, world: world, Animations: animations}, nil
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
	op.GeoM.Translate(p.posX, p.posY)
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
	// Keep player within bounds
	p.posX = max(min(180, p.posX+p.velX), 10)
	p.posY = max(min(105, p.posY+p.velY), 10)
}

func (p *Player) readMovementInputs() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		p.velY = -playerVelocity
		p.orientation = North
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		p.velY = playerVelocity
		p.orientation = South
	} else {
		p.velY = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.velX = -playerVelocity
		p.orientation = West
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		p.velX = playerVelocity
		p.orientation = East
	} else {
		p.velX = 0
	}
}
