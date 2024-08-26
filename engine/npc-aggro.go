package engine

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/jakecoffman/cp"
)

type NpcAggro struct {
	*NpcEntity

	target GameEntity
	// True, once target has entered aggro range
	engaged bool

	waypointInfo WaypointInfo
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
	npc.waypointInfo = opts.WaypointInfo
	npc.wayPoints = opts.WaypointInfo.waypoints
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
		// log.Println("Could not find wp to go to", err.Error())
		// Idle
		body.SetVelocityVector(cp.Vector{})
		if err := n.asset.AnimationController().Loop("idle"); err != nil {
			log.Fatalln("Error looping", err.Error())
		}
		return
	}
	n.moveTowards(body, nextReachableWaypoint)
}

// Utilize Dijkstra pathfinding algorithm to determine where to go
// Returns a vector to the next node on the optimal path from src towards dst
// Node distance is distance between wps & target pos
func (n *NpcAggro) nextWaypointWithDijkstra(srcShape *cp.Shape, dst cp.Vector) (cp.Vector, error) {
	// Early exit: If we can see the dst vector, we dont need dijkstra
	_, visible := calcVisibleDistanceUsingBB(n.shape.Space(), srcShape, dst)
	if visible {
		return dst, nil
	}

	dynamicGraph := n.waypointInfo.graph
	wayPoints := n.waypointInfo.waypoints
	DST_NODE_IDX := len(wayPoints)
	SRC_NODE_IDX := len(wayPoints) + 1

	// Add player node & calculate arcs to all waypoints nodes
	// FIX: Optimize to calculate only once per tick
	// FIX: Should also use BB query here (calcVisibleDistanceUsingBB)
	dynamicGraph.AddEmptyVertex(DST_NODE_IDX)
	for toIdx, toWp := range wayPoints {
		dist, visible := calcVisibleDistance(n.shape.Space(), dst, toWp)
		// Ignore non-visible / path blocked nodes
		if !visible {
			continue
		}
		// Add bidirectional arc
		dynamicGraph.AddArc(DST_NODE_IDX, toIdx, dist)
		dynamicGraph.AddArc(toIdx, DST_NODE_IDX, dist)
	}

	// Add npc node & calculate BIDIRECTIONAL arcs to all wp nodes
	// NOTE: Do not need to calculate arcs to target, because that was already covered in early exit
	// NOTE: Need to consider bounding box dimensions here
	dynamicGraph.AddEmptyVertex(SRC_NODE_IDX)
	visibleWaypoints := 0
	for toIdx, toWp := range wayPoints {
		// Define bounding box edges. Only if path from all edges is clear, wp will be selected
		dist, visible := calcVisibleDistanceUsingBB(n.shape.Space(), srcShape, toWp)
		// Ignore non-visible / path blocked nodes
		if !visible {
			continue
		}
		// Add bidirectional arc
		dynamicGraph.AddArc(SRC_NODE_IDX, toIdx, dist)
		dynamicGraph.AddArc(toIdx, SRC_NODE_IDX, dist)
		visibleWaypoints++
	}
	// If we cannot see any way points (and we cannot see the player which we checked earlier)
	// We're going to have a problem. Dijkstra cannot help us here either
	// Current solution is to perform some random movements, hoping that will unstuck the entity
	if visibleWaypoints == 0 {
		log.Println("NPC stuck! Trying to unstuck with random movement")
		return srcShape.Body().Position().Add(cp.Vector{X: rand.Float64()*10 - 5, Y: rand.Float64()*10 - 5}), nil
	}

	// Apply dijkstra to get path
	bestPath, err := dynamicGraph.Shortest(SRC_NODE_IDX, DST_NODE_IDX)
	if err != nil {
		// Add some debugging information
		// str, strErr := dynamicGraph.Export()
		// if strErr != nil {
		// 	return cp.Vector{}, strErr
		// }
		// _ = str
		// log.Println("Graph data\n", str)
		log.Println("Cannot find shortest path. Trying to unstuck with random movement")
		return srcShape.Body().Position().Add(cp.Vector{X: rand.Float64()*10 - 5, Y: rand.Float64()*10 - 5}), nil
	}
	secondNodeInPathIdx := bestPath.Path[1]
	// First node is always SRC
	// We know node cannot be dst, because we already checked that in the early exit cond
	// Safe to assume that its a waypoint. Checking anyways for good measure
	if secondNodeInPathIdx > len(wayPoints)-1 {
		return cp.Vector{}, fmt.Errorf("Waypoints out of bounds. Received idx %d, but only know %d waypoints", secondNodeInPathIdx, len(wayPoints))
	}
	return wayPoints[secondNodeInPathIdx], nil
}
