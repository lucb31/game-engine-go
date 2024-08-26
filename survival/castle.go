package survival

import (
	"log"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/damage"
	"github.com/lucb31/game-engine-go/engine/hud"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type gameOverCallback = func()
type CastleEntity struct {
	id    engine.GameEntityId
	world engine.GameEntityManager
	engine.GameEntityStats

	// Logic
	health float64

	// Rendering
	asset *engine.CharacterAsset

	// Gun
	gun engine.Gun

	// Physics
	shape *cp.Shape

	gameOverCallback gameOverCallback
}

func NewCastle(world engine.GameEntityManager, cb gameOverCallback) (*CastleEntity, error) {
	c := &CastleEntity{world: world, gameOverCallback: cb, GameEntityStats: engine.DefaultGameEntityStats()}
	// Physical body
	body := cp.NewKinematicBody()
	body.SetVelocityUpdateFunc(c.calculateVelocity)
	body.UserData = c

	// Collision model
	c.shape = cp.NewBox(body, 192, 128, 1)
	c.shape.SetFilter(engine.TowerCollisionFilter())
	c.shape.SetCollisionType(engine.CastleCollision)

	// Register npc collision handler
	handler := world.Space().NewCollisionHandler(engine.CastleCollision, engine.NpcCollision)
	handler.BeginFunc = c.OnCastleHit

	return c, nil
}

func (e *CastleEntity) Draw(screen engine.RenderingTarget) error {
	if e.asset != nil {
		return e.asset.Draw(screen, e.shape, 0)
	} else {
		return engine.DrawRectBoundingBox(screen, e.shape)
	}
}

func (e *CastleEntity) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Automatically shoot
	if e.gun != nil && !e.gun.IsReloading() {
		if err := e.gun.Shoot(); err != nil {
			log.Println("Error when trying to shoot caslte gun", err.Error())
		}
	}
}

func (e *CastleEntity) OnCastleHit(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	_, b := arb.Bodies()
	npc, ok := b.UserData.(damage.Attacker)
	if !ok {
		log.Println("Error in castle on hit: Expected attacker but did not receive one")
		return false
	}
	record, err := e.world.DamageModel().ApplyDamage(npc, e, e.world.IngameTime())
	if err != nil {
		log.Println("Error during castle npc collision damage calc", err.Error())
		return false
	}

	entity, ok := b.UserData.(engine.GameEntity)
	if !ok {
		log.Println("Error during castle npc collision entity removal. Invalid entity provided")
		return false
	}

	log.Println("Castle hit!", record)
	// Remove npc (without loot)
	entity.Destroy()
	return false
}

func (e *CastleEntity) Destroy() error {
	err := e.world.RemoveEntity(e)
	if err != nil {
		return err
	}
	e.gameOverCallback()
	return nil
}
func (e *CastleEntity) Id() engine.GameEntityId               { return e.id }
func (e *CastleEntity) SetId(id engine.GameEntityId)          { e.id = id }
func (e *CastleEntity) Shape() *cp.Shape                      { return e.shape }
func (e *CastleEntity) LootTable() loot.LootTable             { return loot.NewEmptyLootTable() }
func (e *CastleEntity) SetAsset(asset *engine.CharacterAsset) { e.asset = asset }
func (e *CastleEntity) SetGun(gun engine.Gun)                 { e.gun = gun }
func (e *CastleEntity) SetPosition(pos cp.Vector)             { e.shape.Body().SetPosition(pos) }
func (e *CastleEntity) IsVulnerable() bool                    { return true }

func (e *CastleEntity) HealthBar() hud.ProgressInfo {
	return hud.ProgressInfo{0, int(e.MaxHealth()), int(e.Health()), "Castle health"}
}
