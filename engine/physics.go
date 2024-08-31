package engine

import (
	"image/color"
	"log"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

const (
	PlayerCollision cp.CollisionType = iota + 1
	ProjectileCollision
	NpcCollision
	ItemCollision
	CastleCollision
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

// Pass damage model and in game timer to shapes
type SpaceUserData struct {
	damageModel damage.DamageModel
	gameTime    *float64
}

func (s *SpaceUserData) IngameTime() float64 {
	return *s.gameTime
}

func NewPhysicsSpace(damageModel damage.DamageModel, gameTime *float64) (*cp.Space, error) {
	// Initialize physics
	space := cp.NewSpace()
	// Assign references to IGT & damage model to physical space to make
	// them available within every entity
	userData := SpaceUserData{damageModel, gameTime}
	space.StaticBody.UserData = userData
	// NOTE: As long as we're not utilizing collision solvers, we dont need any iterations
	// Therefore setting to min value: 1
	space.Iterations = 1
	// Register Projectile - NPC collision handler
	handler := space.NewCollisionHandler(cp.CollisionType(ProjectileCollision), cp.CollisionType(NpcCollision))
	// TODO: Check if we can replace by static body user data
	handler.UserData = userData
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

func removeProjectile(arb *cp.Arbiter, space *cp.Space, userData interface{}) {
	a, b := arb.Bodies()

	projectile, ok := a.UserData.(*Projectile)
	if !ok {
		log.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return
	}
	collisionPartner, ok := b.UserData.(GameEntity)
	// Ignore collision with projectile owners
	if ok && collisionPartner.Id() == projectile.gun.Owner().Id() {
		return
	}
	log.Printf("Removing projectile after collision with %v \n", b.UserData)
	projectile.Destroy()
}

func projectileCollisionHandler(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	// Validate correct collision partners & type assert
	a, b := arb.Bodies()
	projectile, ok := a.UserData.(*Projectile)
	if !ok {
		log.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return false
	}
	defender, ok := b.UserData.(damage.Defender)
	if !ok {
		log.Println("Type assertion for projectile collision failed. Did not receive valid Damage defender", b.UserData)
		return false
	}

	// Read damage model from userData
	handlerData, ok := userData.(SpaceUserData)
	if !ok {
		log.Println("Could not read damage model")
		return false
	}
	damageModel := handlerData.damageModel
	// Calculate & apply damage with COPY of projectile
	damageResult, err := damageModel.ApplyDamage(projectile, defender, *handlerData.gameTime)
	if err != nil {
		log.Println("Could not apply damage", err.Error())
	}

	// Check if we need to distribute loot
	lootReceiver, isLootReceiver := projectile.gun.Owner().(GameEntityWithInventory)
	if damageResult != nil && damageResult.Fatal && isLootReceiver {
		defenderEntity, isGameEntity := defender.(GameEntity)
		if !isGameEntity {
			log.Println("ERROR: Expected game entity for defender")
			return false
		}
		if err = lootReceiver.Inventory().AddLoot(defenderEntity.LootTable()); err != nil {
			log.Println("Error while adding loot: ", err.Error())
		}
	}

	// Remove projectile
	if err := projectile.OnHit(); err != nil {
		log.Println("Error during projectile on hit handler: %e", err.Error())
		return false
	}
	return false
}

func disableCollisionHandler(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	return false
}

// Draws Rect bounding box around shape position
func DrawRectBoundingBox(t RenderingTarget, bb cp.BB) error {
	topLeft := cp.Vector{bb.L, bb.B}
	botRight := cp.Vector{bb.R, bb.T}

	t.StrokeRect(topLeft, botRight, 2.5, color.RGBA{255, 0, 0, 255}, false)
	return nil
}
