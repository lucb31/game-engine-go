package td

import (
	"fmt"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/hud"
	"github.com/lucb31/game-engine-go/engine/loot"
)

const startingHealth = float64(100.0)
const CastleCollision = cp.CollisionType(200)

type gameOverCallback = func()
type CastleEntity struct {
	id    engine.GameEntityId
	world engine.EntityRemover

	// Logic
	health float64

	// Rendering
	asset     *engine.CharacterAsset
	animation string

	// Physics
	shape *cp.Shape

	gameOverCallback gameOverCallback
}

func NewCastle(world engine.EntityRemover, asset *engine.CharacterAsset, cb gameOverCallback) (*CastleEntity, error) {
	c := &CastleEntity{world: world, asset: asset, health: startingHealth, gameOverCallback: cb}
	c.animation = "idle"
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 640, Y: 395})
	body.SetType(cp.BODY_KINEMATIC)
	body.UserData = c
	c.shape = cp.NewBox(body, 70, 50, 0)
	c.shape.SetFilter(engine.TowerCollisionFilter())
	c.shape.SetCollisionType(CastleCollision)
	return c, nil
}

func (t *CastleEntity) Draw(screen engine.RenderingTarget) error {
	return t.asset.Draw(screen, t.animation, t.shape)
}

func (e *CastleEntity) OnNpcHit(npc *engine.NpcEntity) {
	// TODO: Utilize damage model here
	e.health -= 20
	fmt.Printf("Castle has hit by npc %d. New health %f \n", npc.Id(), e.health)
	if err := npc.Destroy(); err != nil {
		fmt.Println("Could not remove npc", err.Error())
	}
	if e.health <= 0 {
		e.Destroy()
	}
}

func (e *CastleEntity) Destroy() error {
	err := e.world.RemoveEntity(e)
	if err != nil {
		return err
	}
	e.gameOverCallback()
	return nil
}
func (e *CastleEntity) Id() engine.GameEntityId      { return e.id }
func (e *CastleEntity) SetId(id engine.GameEntityId) { e.id = id }
func (e *CastleEntity) Shape() *cp.Shape             { return e.shape }
func (e *CastleEntity) LootTable() loot.LootTable    { return loot.NewEmptyLootTable() }

func (e *CastleEntity) GetHealthBar() hud.ProgressInfo {
	return hud.ProgressInfo{0, int(startingHealth), int(e.health), "Castle health"}
}
