package engine

import (
	"log"
	"slices"

	"github.com/jakecoffman/cp"
)

type GunTargetController struct {
	Gun
	potentialTargets []ProjectileTarget
}

func newGunTargetController(gun Gun) (*GunTargetController, error) {
	return &GunTargetController{Gun: gun}, nil
}

var gunTargetCollisionFilter = cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, NpcCategory)

// Single target choice by collision query
func (c *GunTargetController) chooseTarget() ProjectileTarget {
	query := c.Owner().Shape().Space().PointQueryNearest(c.Position(), c.FireRange(), gunTargetCollisionFilter)
	if query.Shape == nil {
		return nil
	}
	npc, ok := query.Shape.Body().UserData.(*NpcEntity)
	if !ok {
		log.Println("Expected npc target, but found something else", query.Shape.Body().UserData)
		return nil
	}
	return npc
}

func (c *GunTargetController) multiTargetCollisionHandler(shape *cp.Shape, points *cp.ContactPointSet) {
	// Ignore collisions that dont match the gun collision filter
	rejected := gunTargetCollisionFilter.Reject(shape.Filter)
	if rejected {
		return
	}
	// Ignore collisions on multiple contact points
	if points.Count != 1 {
		log.Println("Error: More than one contact point received. Dont know what to do", points.Count)
		return
	}
	// NOTE: Findings about contact point set:
	// points.Points[0] includes two vectors (a,b). My guess is that they build the contact edge
	// points.Points[1] seems to always be 0,0

	// Track target
	userData := shape.Body().UserData
	npc, ok := userData.(*NpcEntity)
	if !ok {
		log.Println("Expected npc target, but found something else", userData)
		return
	}
	c.potentialTargets = append(c.potentialTargets, npc)
}

// Multi target choice by collision query
func (c *GunTargetController) chooseTargets(count int) []ProjectileTarget {
	// Input validation
	if count < 1 {
		return []ProjectileTarget{}
	}
	// Special case: count = 1 makes the query a lot simpler
	if count == 1 {
		tar := c.chooseTarget()
		if tar != nil {
			return []ProjectileTarget{tar}
		}
		return []ProjectileTarget{}
	}

	// Determine potential targets
	circleBody := cp.NewKinematicBody()
	circleBody.SetPosition(c.Position())
	circleShape := cp.NewCircle(circleBody, c.FireRange(), cp.Vector{})
	c.potentialTargets = []ProjectileTarget{}
	c.Owner().Shape().Space().ShapeQuery(circleShape, c.multiTargetCollisionHandler)

	// Sort by distance to gun ASC
	// FIX: Would be more efficient to calculate distances once during collision CB
	slices.SortFunc(c.potentialTargets, func(a, b ProjectileTarget) int {
		distA := a.Body().Position().DistanceSq(c.Position())
		distB := b.Body().Position().DistanceSq(c.Position())
		return int(distA - distB)
	})

	// Return (bounded) list of targets
	if len(c.potentialTargets) < count {
		return c.potentialTargets
	}
	return c.potentialTargets[:count]
}
