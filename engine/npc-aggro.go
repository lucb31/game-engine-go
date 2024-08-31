package engine

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/jakecoffman/cp"
)

type NpcAggro struct {
	*NpcEntity

	target DefenderEntity
	// True, once target has entered attach range
	attacking bool

	swingTimer   Timeout
	waypointInfo WaypointInfo
}

func NewNpcAggro(target DefenderEntity, asset *CharacterAsset, opts NpcOpts) (*NpcAggro, error) {
	if target == nil {
		return nil, fmt.Errorf("Did not receive target")
	}
	base, err := NewNpc(asset, opts)
	if err != nil {
		return nil, err
	}
	npc := &NpcAggro{NpcEntity: base, target: target}
	npc.Shape().Body().SetVelocityUpdateFunc(npc.aggroMovementAI)
	npc.waypointInfo = opts.WaypointInfo
	npc.wayPoints = opts.WaypointInfo.waypoints

	if npc.swingTimer, err = NewIngameTimeout(npc); err != nil {
		return nil, err
	}
	return npc, nil
}

func (n *NpcAggro) IngameTime() float64 {
	u, ok := n.shape.Space().StaticBody.UserData.(SpaceUserData)
	if !ok {
		log.Println("Could not read ingame time")
		return 0.0
	}
	return u.IngameTime()
}

// Main decision tree for npc AI
func (n *NpcAggro) aggroMovementAI(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	if n.attacking {
		// Stop all movement
		body.SetVelocity(0, 0)
		// Wait for next swing timer
		if !n.swingTimer.Done() {
			return
		}

		// Apply damage
		u, ok := n.shape.Space().StaticBody.UserData.(SpaceUserData)
		if !ok {
			log.Println("ERROR: Could not apply damage: No acces to damage model via space user data possible")
			return
		}
		_, err := u.damageModel.ApplyDamage(n, n.target, n.IngameTime())
		if err != nil {
			log.Println("Error during npc swing damage calc", err.Error())
			return
		}

		// Queue up next swing
		if err := n.asset.AnimationController().Play("attack"); err != nil {
			log.Println("Could not play attack animation: %e", err.Error())
		}
		n.swingTimer.Set(1 / n.AtkSpeed())
		return
	}

	// Once within range of target, we start attacking
	if n.target.Shape().BB().Intersects(n.shape.BB()) {
		body.SetVelocity(0, 0)
		n.attacking = true
		// Loop idle animation (in between swings, if animation allows)
		if err := n.asset.AnimationController().Loop("idle"); err != nil {
			log.Println("Could not play idle animation: %e", err.Error())
		}
		// Play first attack animation
		if err := n.asset.AnimationController().Play("attack"); err != nil {
			log.Println("Could not play attack animation: %e", err.Error())
		}
		// Set timer to apply swing damage
		n.swingTimer.Set(1 / n.AtkSpeed())
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
