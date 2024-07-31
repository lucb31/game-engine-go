package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

// Static prop in world. Collidable, but no movement, no animation
type StaticGameEntity struct {
	id    GameEntityId
	Image *ebiten.Image
	shape *cp.Shape
}

func (p *StaticGameEntity) Id() GameEntityId { return p.id }
func (p *StaticGameEntity) Shape() *cp.Shape { return p.shape }

func (p *StaticGameEntity) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.shape.Body().Position().X, p.shape.Body().Position().Y)
	screen.DrawImage(p.Image, &op)
}

func (p *StaticGameEntity) Destroy() {
	fmt.Println("Missing impl: Destryo static entity")
}
