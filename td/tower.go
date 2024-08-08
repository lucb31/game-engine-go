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
	tower := &TowerEntity{world: world, asset: asset, animation: "idle"}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 70, Y: 70})
	body.SetType(cp.BODY_STATIC)
	body.UserData = tower
	tower.shape = cp.NewBox(body, 16, 16, 0)
	return tower, nil
}

func (t *TowerEntity) Draw(screen *ebiten.Image) {
	t.asset.Draw(screen, t.animation, t.shape.Body().Position())
	// TODO: left off: Draw small rect at collision box to verify body is drawn correctly
	// testImg := image.Rect(int(t.shape.Body().Position().X))
	// screen.DrawImage()
}

func (t *TowerEntity) Destroy() {
	t.world.RemoveEntity(t)
}
func (n *TowerEntity) Id() engine.GameEntityId      { return n.id }
func (n *TowerEntity) SetId(id engine.GameEntityId) { n.id = id }
func (n *TowerEntity) Shape() *cp.Shape             { return n.shape }
