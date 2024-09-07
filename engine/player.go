package engine

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type Orientation uint8

const (
	West Orientation = 1 << iota
	North
)

type Player struct {
	// Dependencies
	id              GameEntityId
	world           *GameWorld
	controller      PlayerController
	asset           *CharacterAsset
	projectileAsset *ProjectileAsset
	inventory       loot.Inventory
	*BuildingInteractionController

	// Physics
	shape *cp.Shape

	// Damage model
	gun Gun
	GameEntityStats

	// Harvesting
	axe HarvestingTool

	// Eyeframes
	eyeframesTimeout Timeout
}

const (
	playerWidth                    = 40
	playerHeight                   = 40
	playerPickupRange              = 30.0
	invulnerableForSecondsAfterHit = 0.5
	// Everything in range of this radius will be fully visible (no fog)
	maxVisibilityRadius = 100.0
	// Everything outside of this radius will be fully foggy
	minVisibilityRadius = 200.0
)

func NewPlayer(world *GameWorld, asset *CharacterAsset, projectileAsset *ProjectileAsset) (*Player, error) {
	// Assigning static id -1 to player object
	p := &Player{id: -1, world: world, asset: asset, projectileAsset: projectileAsset}
	// Init player physics
	playerBody := cp.NewBody(1, cp.INFINITY)
	playerBody.UserData = p
	playerBody.SetVelocityUpdateFunc(p.calculateVelocity)

	// Collision model
	p.shape = cp.NewBox(playerBody, playerWidth, playerHeight, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(PlayerCollision))
	p.shape.SetFilter(PlayerCollisionFilter())

	// Register npc collision handler
	ch := world.Space().NewCollisionHandler(cp.CollisionType(PlayerCollision), cp.CollisionType(NpcCollision))
	ch.BeginFunc = p.OnPlayerHit

	// Init stats
	p.GameEntityStats = DefaultGameEntityStats()
	p.movementSpeed = 150
	p.atkSpeed = 1.0

	var err error
	// Init gun
	// gunOpts := BasicGunOpts{FireRange: 500.0}
	// p.gun, err = NewAutoAimGun(world, p, projectileAsset, gunOpts)
	// if err != nil {
	// 	return nil, err
	// }
	// // Play shooting animation when gun shoots
	// p.gun.SetShootingAnimationCallback(func(f float64, orientation Orientation) {
	// 	p.asset.AnimationController().Play("shoot")
	// })

	// Init input controller
	p.controller, err = NewKeyboardPlayerController(p.asset.AnimationController(), p.world)
	if err != nil {
		return nil, err
	}

	// Init eyeframe timer
	p.eyeframesTimeout, err = NewIngameTimeout(world)
	if err != nil {
		return nil, err
	}

	// Init inventory
	p.inventory, err = loot.NewInventory()
	if err != nil {
		return nil, err
	}

	// Init building controller
	p.BuildingInteractionController, err = NewBuildingInteractionController(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Player) Draw(t RenderingTarget) error {
	if err := p.DrawInteractionHud(t); err != nil {
		return err
	}
	if err := p.DrawPlayerStats(t); err != nil {
		return err
	}
	// Early exit: If inside a building we dont want to draw any assets
	if p.Inside() {
		return nil
	}
	p.asset.DrawHealthbar(t, p.shape, p.health, p.maxHealth)
	// Play death animation loop when dead
	if p.health <= 0 || p.world.gameOver {
		err := p.asset.AnimationController().Loop("dead")
		if err != nil {
			log.Println("could not loop death animation", err.Error())
		}
	}
	return p.asset.Draw(t, p.shape, p.controller.Orientation())
}

// TODO: Needs to move to proper hud
func (p *Player) DrawInteractionHud(t RenderingTarget) error {
	var interactionMessage string
	switch {
	case p.Inside():
		interactionMessage = "Press E to exit building"
	case p.BuildingInRange() != nil:
		interactionMessage = "Press E to enter building"
	case p.ItemInRange() != nil:
		interactionMessage = "Press E to pick up"
	case p.axe.Harvesting():
		interactionMessage = "Harvesting..."
	case p.axe.InRange():
		interactionMessage = "Press E to harvest"
	}
	ebitenutil.DebugPrintAt(t.Screen(), interactionMessage, t.Screen().Bounds().Dx()/2-50, t.Screen().Bounds().Dy()/2+25)
	return nil
}

func (p *Player) DrawPlayerStats(t RenderingTarget) error {
	ebitenutil.DebugPrintAt(t.Screen(), "Player stats", t.Screen().Bounds().Dx()-125, 20)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Power %.2f", p.Power()), t.Screen().Bounds().Dx()-125, 35)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Health %.2f", p.Health()), t.Screen().Bounds().Dx()-125, 50)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Max Health %.2f", p.MaxHealth()), t.Screen().Bounds().Dx()-125, 65)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Speed %.2f", p.MovementSpeed()), t.Screen().Bounds().Dx()-125, 80)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("Armor %.2f", p.Armor()), t.Screen().Bounds().Dx()-125, 95)
	ebitenutil.DebugPrintAt(t.Screen(), fmt.Sprintf("AtkSpeed %.2f", p.AtkSpeed()), t.Screen().Bounds().Dx()-125, 110)
	return nil
}

func (p *Player) Destroy() error {
	// Play dying animation
	err := p.asset.AnimationController().Play("die")
	if err != nil {
		log.Println("Could not play dying animation", err.Error())
	}

	// Trigger game over
	p.world.EndGame()
	return nil
}

func (p *Player) OnPlayerHit(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	_, b := arb.Bodies()
	npc, ok := b.UserData.(*NpcEntity)
	if !ok {
		log.Println("Collsion handler error: Expected npc but did not receive one")
		return false
	}
	_, err := p.world.DamageModel().ApplyDamage(npc, p, p.world.IngameTime())
	if err != nil {
		log.Println("Error during player npc collision damage calc", err.Error())
		return false
	}

	// Play on hit animation
	err = p.asset.AnimationController().Play("hit")
	if err != nil {
		log.Println("Could not play on hit animation", err.Error())
	}

	// Register eyeframe timeout
	p.eyeframesTimeout.Set(invulnerableForSecondsAfterHit)

	// npc.Destroy()
	return false
}

func (p *Player) IsVulnerable() bool {
	// Player invulnerable for a brief period after being hit
	if !p.eyeframesTimeout.Done() {
		return false
	}
	// Invulnerable while inside a building
	if p.Inside() {
		return false
	}
	return true
}

func (p *Player) ItemInRange() *ItemEntity {
	queryInfo := p.shape.Space().PointQueryNearest(p.shape.Body().Position(), playerPickupRange, cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, ItemCategory))
	if queryInfo.Shape != nil {
		item, ok := queryInfo.Shape.Body().UserData.(*ItemEntity)
		if !ok {
			log.Println("Error: Expected item, but received sth else")
		}
		return item
	}
	return nil
}

func (p *Player) ItemPickup(item *ItemEntity) error {
	if err := p.Inventory().AddLoot(item.loot); err != nil {
		return err
	}

	// Remove item sprite
	return item.Destroy()
}

func (p *Player) SetAxe(axe HarvestingTool) {
	p.axe = axe
	p.axe.SetAnimationController(p.asset.AnimationController())
}
func (p *Player) Id() GameEntityId          { return p.id }
func (p *Player) SetId(id GameEntityId)     { p.id = id }
func (p *Player) Shape() *cp.Shape          { return p.shape }
func (p *Player) LootTable() loot.LootTable { return loot.NewEmptyLootTable() }
func (p *Player) Inventory() loot.Inventory { return p.inventory }
func (p *Player) Gun() Gun                  { return p.gun }
func (p *Player) Position() cp.Vector       { return p.shape.Body().Position() }

// Do nothing. Already have world reference
func (p *Player) SetEntityRemover(EntityRemover) {}

// Returns true if interaction has FINISHED
func (p *Player) handleInteraction() bool {
	// Exit building
	if p.Inside() {
		if err := p.Leave(); err != nil {
			log.Println("Could not exit building", err.Error())
			return false
		}
		return true
	}

	// Enter building
	buildingInRange := p.BuildingInRange()
	if buildingInRange != nil {
		if err := p.Enter(buildingInRange); err != nil {
			log.Println("Could not enter building", err.Error())
			return false
		}
		return true
	}

	// Item pickup
	itemInRange := p.ItemInRange()
	if itemInRange != nil {
		if err := p.ItemPickup(itemInRange); err != nil {
			log.Println("Could not pickup item", err.Error())
			return false
		}
		return true
	}

	// Harvesting
	if p.axe.InRange() {
		if err := p.axe.HarvestNearest(); err != nil {
			log.Println("Could not harvest", err.Error())
			return false
		}
		// Done, Reset interaction state
		if !p.axe.Harvesting() {
			return true
		}
		// Stop movement
		p.shape.Body().SetVelocity(0, 0)
		return false
	}
	return false
}

func (p *Player) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Read controller inputs
	p.controller.Update()

	// Update harvesting tools
	p.axe.Update()

	// Check for interaction inputs
	if p.controller.Interacting() {
		interactionDone := p.handleInteraction()
		// Early exit: No other inputs / movement if we're still interacting
		if !interactionDone {
			return
		} else {
			// Reset interaction state
			p.controller.SetInteracting(false)
		}
	}

	// Early exit: When inside a building: No player updates possible other than interaction
	if p.Inside() {
		return
	}

	// Update velocity based on inputs
	velocity := p.controller.CalcVelocity(p.MovementSpeed())
	// If moving, abort all prev interactions
	if velocity.LengthSq() > 0.0 {
		if err := p.axe.Abort(); err != nil {
			log.Println("Could not abort harvest", err.Error())
		}
	}
	body.SetVelocityVector(velocity)
	// If we're moving, we need to update fog of war
	p.world.FogOfWar.DiscoverWithRadius(body.Position(), maxVisibilityRadius, minVisibilityRadius)
}
