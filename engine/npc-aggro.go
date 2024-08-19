package engine

import (
	"fmt"
	"slices"

	"github.com/RyanCarrier/dijkstra/v2"
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

	// Move towards target
	dst := n.target.Shape().Body().Position()
	nextReachableWaypoint, err := n.nextWaypointWithDijkstra(n.shape, dst)
	if err != nil {
		fmt.Println("Could not find wp to go to", err.Error())
		// Idle
		body.SetVelocityVector(cp.Vector{})
		n.animation = calculateWalkingAnimation(body.Velocity(), n.orientation)
		return
	}
	n.moveTowards(body, nextReachableWaypoint)
}

// Utilize cp space query to determine if the path between src & dst is clear
// If clear, will return distance between vectors and true
// if not clear, will return 0 and false
func (n *NpcAggro) calcVisibleDistance(src cp.Vector, dst cp.Vector) (uint64, bool) {
	query := n.Shape().Space().SegmentQueryFirst(src, dst, 0.0, cp.NewShapeFilter(cp.NO_GROUP, NpcCategory, OuterWallsCategory))
	// If path between waypoints is blocked, we skip
	if query.Shape != nil {
		return 0, false
	}
	// Calculate distance between waypoints
	return uint64(src.Distance(dst)), true
}

// Utilize cp space query to determine if the path between src BB & dst vector is clear
// Path is only considered clear, if all 4 edges of the BB are visible
// If clear, will return distance between bounding box CENTER position and true
// if not clear, will return 0 and false
func (n *NpcAggro) calcVisibleDistanceUsingBB(src *cp.Shape, dst cp.Vector) (uint64, bool) {
	edges := []cp.Vector{
		TopLeftBBPosition(n.shape),
		TopRightBBPosition(n.shape),
		BottomLeftBBPosition(n.shape),
		BottomRightBBPosition(n.shape),
	}
	// Check all 4 edges and return 0, false if any of them is not visible
	for _, edge := range edges {
		_, visible := n.calcVisibleDistance(edge, dst)
		if !visible {
			return 0, false
		}
	}
	// Return calc distance for center of BB
	return n.calcVisibleDistance(src.Body().Position(), dst)
}

// Utilize Dijkstra pathfinding algorithm to determine where to go
// Returns a vector to the next node on the optimal path from src towards dst
// Node distance is distance between wps & target pos
func (n *NpcAggro) nextWaypointWithDijkstra(srcShape *cp.Shape, dst cp.Vector) (cp.Vector, error) {
	// Early exit: If we can see the dst vector, we dont need dijkstra
	_, visible := n.calcVisibleDistanceUsingBB(srcShape, dst)
	if visible {
		return dst, nil
	}
	// FIX: Once we split up the parts of building the graph, we need to reconsider how to generate these
	DST_NODE_IDX := len(n.wayPoints)
	SRC_NODE_IDX := len(n.wayPoints) + 1

	// Build graph that connects waypoints between each others
	// FIX: Completely static. Optimize to calculate once per game / map
	staticGraph := dijkstra.NewGraph()
	for fromIdx, fromWp := range n.wayPoints {
		staticGraph.AddEmptyVertex(fromIdx)
		// Determine arcs / node edges
		// Nodes are connected if there is no collision in between => cp query
		for toIdx, toWp := range n.wayPoints {
			// Ignore myself
			if toIdx == fromIdx {
				continue
			}
			dist, visible := n.calcVisibleDistance(fromWp, toWp)
			// Ignore non-visible / path blocked nodes
			if !visible {
				continue
			}
			// Add arc
			staticGraph.AddArc(fromIdx, toIdx, dist)
		}
	}

	// Add player node & calculate arcs to all waypoints nodes
	// FIX: Optimize to calculate only once per tick
	// FIX: Should also use BB query here (calcVisibleDistanceUsingBB)
	staticGraph.AddEmptyVertex(DST_NODE_IDX)
	for toIdx, toWp := range n.wayPoints {
		dist, visible := n.calcVisibleDistance(dst, toWp)
		// Ignore non-visible / path blocked nodes
		if !visible {
			continue
		}
		// Add bidirectional arc
		staticGraph.AddArc(DST_NODE_IDX, toIdx, dist)
		staticGraph.AddArc(toIdx, DST_NODE_IDX, dist)
	}

	// Add npc node & calculate BIDIRECTIONAL arcs to all wp nodes
	// NOTE: Do not need to calculate arcs to target, because that was already covered in early exit
	// NOTE: Need to consider bounding box dimensions here
	staticGraph.AddEmptyVertex(SRC_NODE_IDX)
	for toIdx, toWp := range n.wayPoints {
		// Define bounding box edges. Only if path from all edges is clear, wp will be selected
		dist, visible := n.calcVisibleDistanceUsingBB(srcShape, toWp)
		// Ignore non-visible / path blocked nodes
		if !visible {
			continue
		}
		// Add bidirectional arc
		staticGraph.AddArc(SRC_NODE_IDX, toIdx, dist)
		staticGraph.AddArc(toIdx, SRC_NODE_IDX, dist)
	}

	// Apply dijkstra to get path
	bestPath, err := staticGraph.Shortest(SRC_NODE_IDX, DST_NODE_IDX)
	if err != nil {
		// Add some debugging information
		str, strErr := staticGraph.Export()
		if strErr != nil {
			return cp.Vector{}, strErr
		}
		fmt.Println("Graph data", str)
		return cp.Vector{}, err
	}
	secondNodeInPathIdx := bestPath.Path[1]
	// First node is always SRC
	// We know node cannot be dst, because we already checked that in the early exit cond
	// Safe to assume that its a waypoint. Checking anyways for good measure
	if secondNodeInPathIdx > len(n.wayPoints)-1 {
		return cp.Vector{}, fmt.Errorf("Waypoints out of bounds. Received idx %d, but only know %d waypoints", secondNodeInPathIdx, len(n.wayPoints))
	}
	return n.wayPoints[secondNodeInPathIdx], nil
}

// Pathfinding algorithm that will
// - move towards target entity if it can "see" it ELSE
// - move towards the next visible "aiWaypoint"  ELSE
// - Idle
// DEPRECATED, remove once dijkstra tested thoroughly
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
