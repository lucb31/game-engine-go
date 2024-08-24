package engine

import (
	"testing"
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

func TestGunTargetCollisionFilter(t *testing.T) {
	filter := gunTargetCollisionFilter
	npc := NpcCollisionFilter()
	if filter.Reject(npc) {
		t.Fatal("Collision between query and npc was rejected, but should collide")
	}
	if !filter.Reject(HarvestableCollisionFilter) {
		t.Fatal("Expected collision between gun target scanner and tree to be rejected, but was checked")
	}
}
