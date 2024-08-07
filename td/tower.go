package td

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type TowerEntity struct {
	id    engine.GameEntityId
	world engine.GameEntityManager

	// Rendering
	asset     *engine.CharacterAsset
	animation string

	// Physics
	shape *cp.Shape
}

func NewTower(world engine.GameEntityManager, asset *engine.CharacterAsset) (*TowerEntity, error) {
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 25, Y: 10})
	return &TowerEntity{world: world, asset: asset, animation: "idle"}, nil
}

func (t *TowerEntity) Draw(screen *ebiten.Image) {
	t.asset.Draw(screen, t.animation, t.shape.Body().Position())
}

func (t *TowerEntity) Destroy() {
	t.world.RemoveEntity(t)
}
func (n *TowerEntity) Id() engine.GameEntityId      { return n.id }
func (n *TowerEntity) SetId(id engine.GameEntityId) { n.id = id }
func (n *TowerEntity) Shape() *cp.Shape             { return n.shape }
