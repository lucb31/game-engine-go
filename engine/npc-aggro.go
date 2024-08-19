package engine

import (
	"fmt"
	"slices"

	"github.com/jakecoffman/cp"
)

type NpcAggro struct {
	*NpcEntity

	target GameEntity
	// True, once target has entered aggro range
	engaged bool
}

func NewNpcAggro(remover EntityRemover, target GameEntity, asset *CharacterAsset, opts NpcOpts) (*NpcAggro, error) {
	if target == nil {
		return nil, fmt.Errorf("Did not receive target")
	}
	base, err := NewNpc(remover, asset, opts)
	if err != nil {
		return nil, err
	}
	npc := &NpcAggro{NpcEntity: base, target: target}
	npc.Shape().Body().SetVelocityUpdateFunc(npc.aggroMovementAI)
	return npc, nil
}

func (n *NpcAggro) aggroMovementAI(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// TODO: Engage algorithm
	if !n.engaged {
		n.engaged = true
		return
	}

	n.waypointAlgorithmWithCollisionDetection(body)
}

// Pathfinding algorithm that will
// - move towards target entity if it can "see" it ELSE
// - move towards the next visible "aiWaypoint"  ELSE
// - Idle
func (n *NpcAggro) waypointAlgorithmWithCollisionDetection(body *cp.Body) {
	// Define bounding box edges. Only if path from all edges is clear, wp will be selected
	npcEdgesToCheck := []cp.Vector{
		TopLeftBBPosition(n.shape),
		TopRightBBPosition(n.shape),
		BottomLeftBBPosition(n.shape),
		BottomRightBBPosition(n.shape),
	}
	targetPosition := n.target.Shape().Body().Position()
	// Reorder waypoints by proximity to target
	// This ensures that the npc always tries to reach the target instead of being stuck on wp 0
	slices.SortFunc(n.wayPoints, func(a, b cp.Vector) int {
		return int(a.Distance(targetPosition) - b.Distance(targetPosition))
	})

	// Iterate over all waypoints PLUS trying to move towards target first
	for idx := -1; idx < len(n.wayPoints); idx++ {
		// On first iteration try to move towards target
		var wpPosition cp.Vector
		if idx == -1 {
			wpPosition = targetPosition
		} else {
			wpPosition = n.wayPoints[idx]
		}

		// Check if the path between any BB edge and the target position is blocked by a wall
		pathBlocked := false
		for _, edge := range npcEdgesToCheck {
			// NOTE: No idea what the "radius" attribute of that query method is supposed to do. Results did not change at all
			query := n.Shape().Space().SegmentQueryFirst(edge, wpPosition, 0.0, cp.NewShapeFilter(cp.NO_GROUP, NpcCategory, OuterWallsCategory))
			if query.Shape != nil {
				pathBlocked = true
				break
			}
		}
		// Path is blocked. Try next WP
		if pathBlocked {
			continue
		}

		// Path is clear. Initiate movement
		n.moveTowards(body, wpPosition)
		return
	}

	// Idle
	body.SetVelocityVector(cp.Vector{})
	n.animation = calculateWalkingAnimation(body.Velocity(), n.orientation)
}

// DEBUG: Draw connecting lines between npcs & waypoints
// for idx, wp := range aiWaypoints {
// 	// Draw wp index (top left corner, not centered)
// 	relWpPos := w.camera.AbsToRel(wp)
// 	ebitenutil.DebugPrintAt(w.camera.Screen(), fmt.Sprintf("%d", idx), int(relWpPos.X), int(relWpPos.Y))

// 	// Draw connecting lines to npcs
// 	for _, obj := range w.objects {
// 		if _, ok := obj.Shape().Body().UserData.(*NpcEntity); ok {
// 			topLeftNpc := w.camera.AbsToRel(TopLeftBBPosition(obj.Shape()))
// 			botRightNpc := w.camera.AbsToRel(BottomRightBBPosition(obj.Shape()))
// 			vector.StrokeLine(screen, float32(relWpPos.X), float32(relWpPos.Y), float32(topLeftNpc.X), float32(topLeftNpc.Y), 1.0, color.Black, false)
// 			vector.StrokeLine(screen, float32(relWpPos.X), float32(relWpPos.Y), float32(botRightNpc.X), float32(botRightNpc.Y), 1.0, color.Black, false)
// 		}
// 	}
// }
