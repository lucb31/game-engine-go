package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type CustomCollisionType cp.CollisionType

const (
	PlayerCollision CustomCollisionType = iota + 1
	ProjectileCollision
	NpcCollision
)

const (
	PlayerCategory     uint = 1
	NpcCategory        uint = 1 << 1
	OuterWallsCategory uint = 1 << 2
	InnerWallsCategory uint = 1 << 3
	TowerCategory      uint = 1 << 4
	ProjectileCategory uint = 1 << 5
)

// Used to pass damage model and in game timer to collision callback
type HandlerUserData struct {
	damageModel damage.DamageModel
	gameTime    *float64
}

func NewPhysicsSpace(damageModel damage.DamageModel, gameTime *float64) (*cp.Space, error) {
	// Initialize physics
	space := cp.NewSpace()
	// Register collision handlers
	handler := space.NewCollisionHandler(cp.CollisionType(ProjectileCollision), cp.CollisionType(NpcCollision))
	handler.UserData = HandlerUserData{damageModel, gameTime}
	handler.BeginFunc = projectileCollisionHandler
	space.NewWildcardCollisionHandler(cp.CollisionType(ProjectileCollision)).PostSolveFunc = removeProjectile
	return space, nil
}

func TowerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, TowerCategory, PlayerCategory|NpcCategory|OuterWallsCategory|InnerWallsCategory|TowerCategory)
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
	_, err := damageModel.ApplyDamage(projectile, defender, *handlerData.gameTime)
	if err != nil {
		fmt.Println("Could not apply damage", err.Error())
	}
	// Remove projectile
	projectile.Destroy()
	return false
}
