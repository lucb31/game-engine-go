package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type CustomCollisionType cp.CollisionType

const (
	PlayerCollision CustomCollisionType = iota
	ProjectileCollision
	NpcCollision
)

type CollisionCategory uint

const (
	PlayerCategory CollisionCategory = iota + 1
	NpcCategory
	OuterWallsCategory
	InnerWallsCategory
	TowerCategory
	ProjectileCategory
)

func NewPhysicsSpace() (*cp.Space, error) {
	// Initialize physics
	space := cp.NewSpace()
	// Register collision handlers
	handler := space.NewCollisionHandler(cp.CollisionType(ProjectileCollision), cp.CollisionType(NpcCollision))
	handler.BeginFunc = npcProjectilecollisionHandler
	space.NewWildcardCollisionHandler(cp.CollisionType(ProjectileCollision)).PostSolveFunc = removeProjectile
	return space, nil
}

func TowerCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, uint(TowerCategory), uint(PlayerCategory|NpcCategory|OuterWallsCategory|InnerWallsCategory))
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
	if ok && collisionPartner.Id() == projectile.owner.Id() {
		return
	}
	fmt.Printf("Removing projectile after collision with %v \n", b.UserData)
	projectile.Destroy()
}

func npcProjectilecollisionHandler(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	// Validate correct collision partners & type assert
	a, b := arb.Bodies()
	projectile, ok := a.UserData.(*Projectile)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return false
	}
	npc, ok := b.UserData.(*NpcEntity)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Npc", b.UserData)
		return false
	}
	// Trigger projectile hit with COPY of projectile
	npc.OnProjectileHit(*projectile)
	// Remove projectile
	projectile.Destroy()
	return false
}
