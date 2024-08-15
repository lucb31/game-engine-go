package td

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

const startingHealth = float64(100.0)
const CastleCollision = cp.CollisionType(200)

type CastleEntity struct {
	id    engine.GameEntityId
	world engine.GameEntityManager

	// Logic
	health float64

	// Rendering
	asset     *engine.CharacterAsset
	animation string

	// Physics
	shape *cp.Shape
}

func NewCastle(world engine.GameEntityManager, asset *engine.CharacterAsset) (*CastleEntity, error) {
	c := &CastleEntity{world: world, asset: asset, health: startingHealth}
	c.animation = "idle"
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 640, Y: 384})
	body.SetType(cp.BODY_KINEMATIC)
	body.UserData = c
	c.shape = cp.NewBox(body, 32, 32, 0)
	c.shape.SetFilter(engine.TowerCollisionFilter())
	c.shape.SetCollisionType(CastleCollision)
	return c, nil
}

func (t *CastleEntity) Draw(screen *ebiten.Image) {
	t.asset.Draw(screen, t.animation, t.shape.Body().Position())
}

func (e *CastleEntity) OnNpcHit(npc *engine.NpcEntity) {
	e.health -= 20
	fmt.Printf("Castle has hit by npc %d. New health %f \n", npc.Id(), e.health)
	npc.Destroy()
	if e.health <= 0 {
		e.Destroy()
	}
}

func (e *CastleEntity) Destroy() error {
	err := e.world.RemoveEntity(e)
	if err != nil {
		return err
	}
	e.world.EndGame()
	return nil
}
func (e *CastleEntity) Id() engine.GameEntityId      { return e.id }
func (e *CastleEntity) SetId(id engine.GameEntityId) { e.id = id }
func (e *CastleEntity) Shape() *cp.Shape             { return e.shape }

func (e *CastleEntity) GetHealthBar() ProgressInfo {
	return ProgressInfo{0, int(startingHealth), int(e.health), "Castle health"}
}
