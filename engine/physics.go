package engine

import (
	"fmt"
	"image/color"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

const (
	PlayerCollision cp.CollisionType = iota + 1
	ProjectileCollision
	NpcCollision
	ItemCollision
)

const (
	PlayerCategory      uint = 1 << iota
	NpcCategory         uint = 1 << iota
	OuterWallsCategory  uint = 1 << iota
	HarvestableCategory uint = 1 << iota
	TowerCategory       uint = 1 << iota
	ProjectileCategory  uint = 1 << iota
	ItemCategory        uint = 1 << iota
)

// Used to pass damage model and in game timer to collision callback
type HandlerUserData struct {
	damageModel damage.DamageModel
	gameTime    *float64
}

func NewPhysicsSpace(damageModel damage.DamageModel, gameTime *float64) (*cp.Space, error) {
	// Initialize physics
	space := cp.NewSpace()
	// Register Projectile - NPC collision handler
	handler := space.NewCollisionHandler(cp.CollisionType(ProjectileCollision), cp.CollisionType(NpcCollision))
	handler.UserData = HandlerUserData{damageModel, gameTime}
	handler.BeginFunc = projectileCollisionHandler

	// Disable all collision with items
	itemHandler := space.NewWildcardCollisionHandler(ItemCollision)
	itemHandler.BeginFunc = disableCollisionHandler

	// Register projectile wildcard handler to remove stray projectiles
	space.NewWildcardCollisionHandler(cp.CollisionType(ProjectileCollision)).PostSolveFunc = removeProjectile
	return space, nil
}

func TowerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, TowerCategory, PlayerCategory|NpcCategory|OuterWallsCategory|HarvestableCategory|TowerCategory)
}

func PlayerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, PlayerCategory, PlayerCategory|NpcCategory|OuterWallsCategory|TowerCategory|HarvestableCategory)
}

func TopLeftBBPosition(shape *cp.Shape) cp.Vector {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	return cp.Vector{
		X: shape.Body().Position().X - width/2,
		Y: shape.Body().Position().Y - height/2,
	}
}

func TopRightBBPosition(shape *cp.Shape) cp.Vector {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	return cp.Vector{
		X: shape.Body().Position().X + width/2,
		Y: shape.Body().Position().Y - height/2,
	}
}

func BottomLeftBBPosition(shape *cp.Shape) cp.Vector {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	return cp.Vector{
		X: shape.Body().Position().X - width/2,
		Y: shape.Body().Position().Y + height/2,
	}
}

func BottomRightBBPosition(shape *cp.Shape) cp.Vector {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	return cp.Vector{
		X: shape.Body().Position().X + width/2,
		Y: shape.Body().Position().Y + height/2,
	}
}

func removeProjectile(arb *cp.Arbiter, space *cp.Space, userData interface{}) {
	a, b := arb.Bodies()

	projectile, ok := a.UserData.(*Projectile)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return
	}
	collisionPartner, ok := b.UserData.(GameEntity)
	// Ignore collision with projectile owners
	if ok && collisionPartner.Id() == projectile.gun.Owner().Id() {
		return
	}
	fmt.Printf("Removing projectile after collision with %v \n", b.UserData)
	projectile.Destroy()
}

func projectileCollisionHandler(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	// Validate correct collision partners & type assert
	a, b := arb.Bodies()
	projectile, ok := a.UserData.(*Projectile)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return false
	}
	defender, ok := b.UserData.(damage.Defender)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Damage defender", b.UserData)
		return false
	}

	// Read damage model from userData
	handlerData, ok := userData.(HandlerUserData)
	if !ok {
		fmt.Println("Could not read damage model")
		return false
	}
	damageModel := handlerData.damageModel
	// Calculate & apply damage with COPY of projectile
	damageResult, err := damageModel.ApplyDamage(projectile, defender, *handlerData.gameTime)
	if err != nil {
		fmt.Println("Could not apply damage", err.Error())
	}

	// Check if we need to distribute loot
	lootReceiver, isLootReceiver := projectile.gun.Owner().(GameEntityWithInventory)
	if damageResult.Fatal && isLootReceiver {
		defenderEntity, isGameEntity := defender.(GameEntity)
		if !isGameEntity {
			fmt.Println("ERROR: Expected game entity for defender")
			return false
		}
		if err = lootReceiver.Inventory().AddLoot(defenderEntity.LootTable()); err != nil {
			fmt.Println("Error while adding loot: ", err.Error())
		}
	}

	// Remove projectile
	projectile.Destroy()
	return false
}

func disableCollisionHandler(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	return false
}

// Draws Rect bounding box around shape position
func DrawRectBoundingBox(t RenderingTarget, shape *cp.Shape) error {
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	t.StrokeRect(shape.Body().Position().X-width/2, shape.Body().Position().Y-height/2, float32(width), float32(height), 2.5, color.RGBA{255, 0, 0, 255}, false)
	return nil
}
