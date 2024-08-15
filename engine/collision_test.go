package engine

import (
	"testing"

	"github.com/jakecoffman/cp"
)

func TestPlayerCollisionFilters(t *testing.T) {
	player := PlayerCollisionFilter()
	tower := TowerCollisionFilter()
	outerWall := BoundingBoxFilter()
	npc := NpcCollisionFilter()
	projectile := ProjectileCollisionFilter()
	if player.Reject(player) {
		t.Fatal("Collision between player and player was ignored")
	}
	if player.Reject(tower) {
		t.Fatal("Collision between player and tower was rejected. But should collide")
		return
	}
	if player.Reject(outerWall) {
		t.Fatal("Collision between player and outer wall was rejected. But should collide")
	}
	if player.Reject(npc) {
		t.Fatal("Collision between player and npc was rejected. But should collide")
	}
	if !player.Reject(projectile) {
		t.Fatal("Collision between player and projectile was NOT rejected. But should NOT collide")
	}
}

func TestNpcCollisionFilters(t *testing.T) {
	player := PlayerCollisionFilter()
	tower := TowerCollisionFilter()
	outerWall := BoundingBoxFilter()
	npc := NpcCollisionFilter()
	projectile := ProjectileCollisionFilter()
	if npc.Reject(outerWall) {
		t.Fatal("Collision between npc and outer wall was rejected. But should collide")
	}
	if npc.Reject(tower) {
		t.Fatal("Collision between npc and tower was rejected. But should collide")
	}
	if npc.Reject(projectile) {
		t.Fatal("Collision between npc and projectile was rejected. But should collide")
	}
	if npc.Reject(player) {
		t.Fatal("Collision between npc and player was rejected. But should collide")
	}
	if !npc.Reject(npc) {
		t.Fatal("Collision between npc and npc was not rejected. But should ignore collision")
	}
}

func TestTowerCollisionFilter(t *testing.T) {
	player := PlayerCollisionFilter()
	tower := TowerCollisionFilter()
	projectile := ProjectileCollisionFilter()
	if tower.Reject(player) {
		t.Fatal("Collision between tower and player was rejected. But should collide")
	}
	if !tower.Reject(projectile) {
		t.Fatal("Collision between tower and projectile was checked. But should NOT collide")
	}
}

func TestProjectileCollisionFilter(t *testing.T) {
	projectile := ProjectileCollisionFilter()
	if !projectile.Reject(projectile) {
		t.Fatal("Collision between projectile and projectile was checked, but should NOT collide")
	}
}

func TestPointQueryFilter(t *testing.T) {
	filter := cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, NpcCategory)
	npc := NpcCollisionFilter()
	if filter.Reject(npc) {
		t.Fatal("Collision between query and npc was rejected, but should collide")
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
