package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type NpcEntity struct {
	Shape          *cp.Shape
	wayPoints      []cp.Vector
	currentWpIndex int
	loopWaypoints  bool
	asset          *CharacterAsset
	velocity       float64
}

func NewNpc(asset *CharacterAsset) (*NpcEntity, error) {
	npc := &NpcEntity{}
	// Init body & shape
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 50, Y: 50})
	body.SetVelocityUpdateFunc(npc.calculateVelocity)
	npc.Shape = cp.NewBox(body, 8, 8, 0)
	npc.Shape.SetElasticity(0)
	npc.Shape.SetFriction(1)
	npc.asset = asset
	npc.wayPoints = []cp.Vector{
		{X: 20, Y: 20},
		{X: 100, Y: 20},
		{X: 100, Y: 100},
		{X: 20, Y: 100},
	}
	npc.loopWaypoints = true
	npc.velocity = 50.0
	return npc, nil
}

func (n *NpcEntity) calculateOrientation() Orientation {
	vel := n.Shape.Body().Velocity()
	if vel.Y > 5 {
		return South
	} else if vel.Y < -5 {
		return North
	}
	if vel.X > 5 {
		return East
	} else if vel.X < -5 {
		return West
	}
	return East
}

func (n *NpcEntity) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(n.Shape.Body().Position().X, n.Shape.Body().Position().Y)

	// TODO: Refactor: Animation logic duplicated in player
	var animation string
	switch n.calculateOrientation() {
	case North:
		animation = "walk_north"
	case South:
		animation = "walk_south"
	case East:
		animation = "walk_east"
	case West:
		animation = "walk_west"
	default:
		animation = "idle"
	}

	im, err := n.asset.GetTile(animation)
	if err != nil {
		fmt.Println("Could not draw npc", err.Error())
		return
	}
	screen.DrawImage(im, &op)
}

func (n *NpcEntity) Update() {}

// Calculate velocity based on simple pathfinding algorithm between waypoints
func (n *NpcEntity) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// No movement if no active wayPoint
	if n.currentWpIndex == -1 {
		body.SetVelocityVector(cp.Vector{})
		return
	}
	destination := n.wayPoints[n.currentWpIndex]
	position := body.Position()
	diff := destination.Sub(position)
	diffNormalized := diff.Normalize()

	// Go to next waypoint if in close proximity to current WP
	if diff.Length() < 5 {
		fmt.Printf("Waypoint %d reached \n", n.currentWpIndex)
		n.currentWpIndex++
		if n.currentWpIndex > len(n.wayPoints)-1 {
			if n.loopWaypoints {
				n.currentWpIndex = 0
				fmt.Println("Looping")
			} else {
				n.currentWpIndex = -1
				fmt.Println("Stopping movement")
			}
		}
	}
	vel := diffNormalized.Mult(n.velocity)
	body.SetVelocityVector(vel)
}
