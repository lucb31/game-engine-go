package engine

import (
	"github.com/RyanCarrier/dijkstra/v2"
	"github.com/jakecoffman/cp"
)

type WaypointInfo struct {
	waypoints []cp.Vector
	graph     dijkstra.Graph
}

func NewWaypointInfo(space *cp.Space, wps []cp.Vector) (*WaypointInfo, error) {
	return &WaypointInfo{wps, buildStaticWpGraph(space, wps)}, nil
}

// Build graph that connects waypoints between each others
func buildStaticWpGraph(space *cp.Space, wps []cp.Vector) dijkstra.Graph {
	staticGraph := dijkstra.NewGraph()
	for fromIdx, fromWp := range wps {
		staticGraph.AddEmptyVertex(fromIdx)
		// Determine arcs / node edges
		// Nodes are connected if there is no collision in between => cp query
		for toIdx, toWp := range wps {
			// Ignore myself
			if toIdx == fromIdx {
				continue
			}
			dist, visible := calcVisibleDistance(space, fromWp, toWp)
			// Ignore non-visible / path blocked nodes
			if !visible {
				continue
			}
			// Add arc
			staticGraph.AddArc(fromIdx, toIdx, dist)
		}
	}
	return staticGraph
}

// Utilize cp space query to determine if the path between src & dst is clear
// If clear, will return distance between vectors and true
// if not clear, will return 0 and false
func calcVisibleDistance(space *cp.Space, src cp.Vector, dst cp.Vector) (uint64, bool) {
	query := space.SegmentQueryFirst(src, dst, 0.0, cp.NewShapeFilter(cp.NO_GROUP, NpcCategory, OuterWallsCategory))
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
func calcVisibleDistanceUsingBB(space *cp.Space, src *cp.Shape, dst cp.Vector) (uint64, bool) {
	edges := []cp.Vector{
		TopLeftBBPosition(src),
		TopRightBBPosition(src),
		BottomLeftBBPosition(src),
		BottomRightBBPosition(src),
	}
	// Check all 4 edges and return 0, false if any of them is not visible
	for _, edge := range edges {
		_, visible := calcVisibleDistance(space, edge, dst)
		if !visible {
			return 0, false
		}
	}
	// Return calc distance for center of BB
	return calcVisibleDistance(space, src.Body().Position(), dst)
}
