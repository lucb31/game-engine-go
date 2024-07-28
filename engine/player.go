package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	posX  int
	posY  int
	world *GameWorld
}

const (
	frameWidth          = 48
	frameHeight         = 48
	animationFrameCount = 6
)

func NewPlayer(world *GameWorld) (*Player, error) {
	return &Player{posX: 0, posY: 0, world: world}, nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	// Position in the center of the screen
	op := ebiten.DrawImageOptions{}
	// Offset size of player frame
	op.GeoM.Translate(-frameWidth/2, -frameHeight/2)
	op.GeoM.Translate(16*6, 16*4)

	animationFrame := int(p.world.FrameCount/6) % animationFrameCount
	tileRow := 4
	subIm := p.world.AssetManager.PlayerTileset.SubImage(image.Rect(
		animationFrame*frameWidth,
		tileRow*frameHeight,
		(animationFrame+1)*frameWidth,
		(tileRow+1)*frameHeight,
	)).(*ebiten.Image)
	screen.DrawImage(subIm, &op)
}
