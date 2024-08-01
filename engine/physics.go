package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
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

func removeProjectile(arb *cp.Arbiter, space *cp.Space, userData interface{}) {
	p, _ := arb.Bodies()

	projectile, ok := p.UserData.(*Projectile)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return
	}
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
