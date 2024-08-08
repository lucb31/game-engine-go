package main

import (
	"fmt"
	"testing"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

func TestCollisionFilters(t *testing.T) {
	player := engine.PlayerCollisionFilter()
	tower := engine.TowerCollisionFilter()
	outerWall := engine.BoundingBoxFilter()
	npc := engine.NpcCollisionFilter()
	projectile := cp.NewShapeFilter(0, uint(engine.ProjectileCategory), uint(engine.NpcCategory|engine.OuterWallsCategory&^engine.PlayerCategory))
	//	projectile := engine.ProjectileCollisionFilter()
	fmt.Println("Player: ", player.Categories, player.Mask)
	fmt.Println("Tower: ", tower.Categories, tower.Mask)
	fmt.Println("Outer wall: ", outerWall.Categories, outerWall.Mask)
	fmt.Println("Projectile: ", projectile.Categories, projectile.Mask)
	if player.Reject(player) {
		t.Fatal("Collision between player and player was ignored")
	}
	if player.Reject(tower) {
		fmt.Println(player.Group != 0 && player.Group == tower.Group)
		// One of the category/mask combinations fails.
		fmt.Println((player.Categories & tower.Mask) == 0)
		fmt.Println((tower.Categories & player.Mask) == 0)
		t.Fatal("Collision between player and tower was rejected. But should collide")
		return
	}
	if player.Reject(outerWall) {
		fmt.Println(player.Group != 0 && player.Group == outerWall.Group)
		fmt.Println((player.Categories & outerWall.Mask) == 0)
		fmt.Println((outerWall.Categories & player.Mask) == 0)
		t.Fatal("Collision between player and outer wall was rejected. But should collide")
	}
	if player.Reject(npc) {
		t.Fatal("Collision between player and npc was rejected. But should collide")
	}
	if !player.Reject(projectile) {
		fmt.Println(player.Group != 0 && player.Group == projectile.Group)
		fmt.Println((player.Categories & projectile.Mask) == 0)
		fmt.Println((player.Categories & projectile.Mask))
		fmt.Println((projectile.Categories & player.Mask) == 0)
		fmt.Println((projectile.Categories & player.Mask))
		t.Fatal("Collision between player and projectile was NOT rejected. But should NOT collide")
	}
	if npc.Reject(outerWall) {
		t.Fatal("Collision between npc and outer wall was rejected. But should collide")
	}
	if npc.Reject(tower) {
		t.Fatal("Collision between npc and tower was rejected. But should collide")
	}
	if npc.Reject(projectile) {
		t.Fatal("Collision between npc and projectile was rejected. But should collide")
	}
	if outerWall.Reject(tower) {
		t.Fatal("Collision between outer wall and tower was rejected. But should collide")
	}
	if outerWall.Reject(projectile) {
		t.Fatal("Collision between outer wall and projectile was rejected. But should collide")
	}
	if !tower.Reject(projectile) {
		t.Fatal("Collision between tower and projectile was checked. But should NOT collide")
	}
}

type ShapeFilter struct {
	/// Two objects with the same non-zero group value do not collide.
	/// This is generally used to group objects in a composite object together to disable self collisions.
	Group uint
	/// A bitmask of user definable categories that this object belongs to.
	/// The category/mask combinations of both objects in a collision must agree for a collision to occur.
	Categories uint
	/// A bitmask of user definable category types that this object object collides with.
	/// The category/mask combinations of both objects in a collision must agree for a collision to occur.
	Mask uint
}

// Just for reference
func (a ShapeFilter) Reject(b ShapeFilter) bool {
	// Reject the collision if:
	return (a.Group != 0 && a.Group == b.Group) ||
		// One of the category/mask combinations fails.
		(a.Categories&b.Mask) == 0 ||
		(b.Categories&a.Mask) == 0
}
