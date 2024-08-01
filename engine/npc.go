package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type NpcEntity struct {
	id             GameEntityId
	world          *GameWorld
	shape          *cp.Shape
	wayPoints      []cp.Vector
	currentWpIndex int
	loopWaypoints  bool
	asset          *CharacterAsset
	velocity       float64
	animation      string
	orientation    Orientation
}

func NewNpc(world *GameWorld, asset *CharacterAsset) (*NpcEntity, error) {
	npc := &NpcEntity{world: world, orientation: South}
	// Init body & shape
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 50, Y: 50})
	body.SetVelocityUpdateFunc(npc.calculateVelocity)
	body.UserData = npc
	npc.shape = cp.NewBox(body, 8, 8, 0)
	npc.shape.SetElasticity(0)
	npc.shape.SetFriction(1)
	npc.shape.SetCollisionType(cp.CollisionType(NpcCollision))
	npc.asset = asset
	npc.wayPoints = []cp.Vector{
		{X: 20, Y: 20},
		{X: 100, Y: 20},
		{X: 100, Y: 100},
		{X: 20, Y: 100},
	}
	npc.loopWaypoints = true
	npc.velocity = 50.0
	npc.animation = "idle_south"
	return npc, nil
}

func (n *NpcEntity) Draw(screen *ebiten.Image) {
	n.asset.Draw(screen, n.animation, n.shape.Body().Position())
}

func (n *NpcEntity) Destroy() {
	fmt.Println("ERROR: Missing implementation for npc destroy")
}

func (n *NpcEntity) Id() GameEntityId      { return n.id }
func (n *NpcEntity) SetId(id GameEntityId) { n.id = id }
func (n *NpcEntity) Shape() *cp.Shape      { return n.shape }

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
	// Update active animation & orientation
	n.orientation = calculateOrientation(vel)
	n.animation = calculateWalkingAnimation(vel, n.orientation)
}
