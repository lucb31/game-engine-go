package survival

import (
	"fmt"
	"image/color"
	"log"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/damage"
	"github.com/lucb31/game-engine-go/engine/hud"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type gameOverCallback = func()
type CastleEntity struct {
	id     engine.GameEntityId
	world  engine.GameEntityManager
	camera *engine.FollowingCamera
	engine.GameEntityStats

	// Rendering
	asset *engine.CharacterAsset

	// Logic
	playerInside *engine.Player
	// Keep track of last player inside even after leaving. This is required to
	// add loot to player inventory even after he leaves
	recentPlayerInside *engine.Player
	gun                engine.Gun

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
	c.shape = cp.NewCircle(body, 65.0, cp.Vector{})
	c.shape.SetFilter(engine.TowerCollisionFilter())
	c.shape.SetCollisionType(engine.CastleCollision)

	// Register npc collision handler
	handler := world.Space().NewCollisionHandler(engine.CastleCollision, engine.NpcCollision)
	handler.BeginFunc = c.OnCastleHit

	c.SetMaxHealth(5000)
	c.SetHealth(5000)

	return c, nil
}

func (e *CastleEntity) Draw(screen engine.RenderingTarget) error {
	// Draw castle asset or BB
	if e.asset != nil {
		if err := e.asset.Draw(screen, e.shape, 0); err != nil {
			return err
		}
	} else {
		if err := engine.DrawRectBoundingBox(screen, e.shape.BB()); err != nil {
			return err
		}
	}

	// Draw firing range, if gun available
	gun := e.Gun()
	if gun != nil {
		screen.StrokeCircle(e.shape.Body().Position(), float32(gun.FireRange()), 2.0, color.NRGBA{255, 0, 0, 255}, false)
	}
	return nil
}

func (e *CastleEntity) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Automatically shoot
	gun := e.Gun()
	if gun != nil && !gun.IsReloading() {
		if err := gun.Shoot(); err != nil {
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

func (e *CastleEntity) Enter(p engine.GameEntityEntering) error {
	if e.playerInside != nil {
		return fmt.Errorf("Cannot enter: Already a player inside")
	}
	// Cast to player
	player, ok := p.(*engine.Player)
	if !ok {
		return fmt.Errorf("Can only be entered by players")
	}
	e.playerInside = player
	e.recentPlayerInside = player

	// Pan camera
	if e.camera == nil {
		return fmt.Errorf("Could not refocus camera to castle. No camera provided")
	}
	e.camera.SetTarget(e)
	return nil
}

func (e *CastleEntity) Leave(p engine.GameEntityEntering) error {
	if e.playerInside == nil {
		return fmt.Errorf("Cannot leave. No player inside")
	}
	if e.playerInside != p {
		return fmt.Errorf("Cannot leave. Different player inside")
	}

	// Pan camera
	if e.camera == nil {
		return fmt.Errorf("Could not refocus camera to player. No camera provided")
	}
	e.camera.SetTarget(e.playerInside)

	e.playerInside = nil

	return nil
}

func (e *CastleEntity) Gun() engine.Gun {
	// If no player inside, we dont use any guns
	if e.playerInside == nil {
		return nil
	}
	// Return own gun if available
	if e.gun != nil {
		return e.gun
	}
	// Fallback to player gun
	return e.playerInside.Gun()
}

func (e *CastleEntity) SetGun(gun engine.Gun) { e.gun = gun }

func (e *CastleEntity) Id() engine.GameEntityId      { return e.id }
func (e *CastleEntity) SetId(id engine.GameEntityId) { e.id = id }

// Do nothing. We already have a reference to the world
func (e *CastleEntity) SetEntityRemover(engine.EntityRemover) {}
func (e *CastleEntity) Shape() *cp.Shape                      { return e.shape }
func (e *CastleEntity) LootTable() loot.LootTable             { return loot.NewEmptyLootTable() }
func (e *CastleEntity) SetAsset(asset *engine.CharacterAsset) { e.asset = asset }
func (e *CastleEntity) IsVulnerable() bool                    { return true }
func (e *CastleEntity) ShopEnabled() bool                     { return e.playerInside != nil }
func (e *CastleEntity) SetCamera(cam *engine.FollowingCamera) { e.camera = cam }
func (e *CastleEntity) Position() cp.Vector                   { return e.shape.Body().Position() }
func (e *CastleEntity) Inventory() loot.Inventory {
	if e.playerInside != nil {
		return e.playerInside.Inventory()
	} else if e.recentPlayerInside != nil {
		return e.recentPlayerInside.Inventory()
	}
	log.Println("Error: Cannot access player inventory")
	return nil
}

func (e *CastleEntity) HealthBar() hud.ProgressInfo {
	return hud.ProgressInfo{0, int(e.MaxHealth()), int(e.Health()), "Castle health"}
}
