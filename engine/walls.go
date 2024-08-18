package engine

import (
	"math"

	"github.com/jakecoffman/cp"
)

type WallSegment struct {
	Start, End cp.Vector
}

// Scan map data for wall segments
func CalcHorizontalWallSegments(tileData [][]MapTile) []WallSegment {
	horizontalSegments := []WallSegment{}
	currentSegment := WallSegment{}
	for row, rowData := range tileData {
		for col, cellData := range rowData {
			if cellData == EmptyTile {
				// If empty cell, end current segment
				if currentSegment.Start.Length() > 0 {
					// Set end to bottom left world coordinates (left because we're at the next tile here)
					x, y := GridPosToWorldPos(col, row)
					currentSegment.End = cp.Vector{x, y + mapTileSize}
					// Ignore 1 tile segments
					// NOTE: Perpendicular: mapTileSize^2 + mapTileSize^2 = dist^2
					dist := currentSegment.Start.DistanceSq(currentSegment.End)
					if dist > 2*math.Pow(mapTileSize, 2) {
						horizontalSegments = append(horizontalSegments, currentSegment)
					}
					currentSegment = WallSegment{}
				}
			} else {
				// If occupied cell and no active segment, start new segment
				if currentSegment.Start.Length() == 0 {
					// Set start to top left world coordinates
					x, y := GridPosToWorldPos(col, row)
					currentSegment.Start = cp.Vector{x, y}
				}
			}
		}
	}
	return horizontalSegments
}

// Scan map data vertically for wall segments
func CalcVerticalWallSegments(tileData [][]MapTile) []WallSegment {
	verticalSegments := []WallSegment{}
	currentSegment := WallSegment{}
	for col := range len(tileData[0]) {
		for row, rowData := range tileData {
			cellData := rowData[col]
			if cellData == EmptyTile {
				// If empty cell, end current segment
				if currentSegment.Start.Length() > 0 {
					// Set end to top right world coordinates (right because we're at the next tile here)
					x, y := GridPosToWorldPos(col, row)
					currentSegment.End = cp.Vector{x + mapTileSize, y}
					// Ignore 1 tile segments
					// NOTE: Perpendicular: mapTileSize^2 + mapTileSize^2 = dist^2
					dist := currentSegment.Start.DistanceSq(currentSegment.End)
					if dist > 2*math.Pow(mapTileSize, 2) {
						verticalSegments = append(verticalSegments, currentSegment)
					}
					currentSegment = WallSegment{}
				}
			} else {
				// If occupied cell and no active segment, start new segment
				if currentSegment.Start.Length() == 0 {
					// Set start to top left world coordinates
					x, y := GridPosToWorldPos(col, row)
					currentSegment.Start = cp.Vector{x, y}
				}
			}
		}
	}
	return verticalSegments
}

func RegisterWallSegmentToSpace(space *cp.Space, segment WallSegment) {
	shape := space.AddShape(cp.NewSegment(space.StaticBody, segment.Start, segment.End, 2))
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(BoundingBoxFilter())
}

// Returns TOP LEFT position of tile in world coordinate system
func GridPosToWorldPos(col, row int) (float64, float64) {
	return float64(col * mapTileSize), float64(row * mapTileSize)
}
