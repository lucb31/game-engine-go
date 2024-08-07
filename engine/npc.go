package engine

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type NpcEntity struct {
	id    GameEntityId
	world GameEntityManager

	// Logic
	health float64

	// Rendering
	asset       *CharacterAsset
	animation   string
	orientation Orientation

	// Physics
	shape    *cp.Shape
	velocity float64

	// Movement AI
	wayPoints         []cp.Vector
	currentWpIndex    int
	loopWaypoints     bool
	stopMovementUntil time.Time
}

func NewNpc(world GameEntityManager, asset *CharacterAsset) (*NpcEntity, error) {
	npc := &NpcEntity{world: world, orientation: South, health: 100.0}
	// Init body & shape
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 25, Y: 10})
	body.SetVelocityUpdateFunc(npc.calculateVelocity)
	body.UserData = npc
	npc.shape = cp.NewBox(body, 8, 8, 0)
	npc.shape.SetElasticity(0)
	npc.shape.SetFriction(1)
	npc.shape.SetCollisionType(cp.CollisionType(NpcCollision))
	npc.asset = asset
	npc.wayPoints = []cp.Vector{
		{X: 15, Y: 15},
		{X: 15, Y: 340},
		{X: 165, Y: 340},
		{X: 165, Y: 15},
		{X: 310, Y: 15},
		{X: 310, Y: 340},
		{X: 450, Y: 340},
		{X: 450, Y: 10},
	}
	npc.loopWaypoints = false
	npc.velocity = 50.0
	npc.animation = "idle_south"
	return npc, nil
}

func (n *NpcEntity) Draw(screen *ebiten.Image) {
	n.asset.Draw(screen, n.animation, n.shape.Body().Position())
}

func (n *NpcEntity) Destroy() {
	n.world.RemoveEntity(n)
}

func (n *NpcEntity) OnProjectileHit(projectile Projectile) {
	fmt.Println("OUCH!", n, projectile)
	n.health -= 30.0
	if n.health <= 0.0 {
		n.Destroy()
	}
	fmt.Printf("Survived with %f health \n", n.health)
	// Briefly stop movement
	n.stopMovementUntil = time.Now().Add(time.Millisecond * 300)
}

func (n *NpcEntity) Id() GameEntityId      { return n.id }
func (n *NpcEntity) SetId(id GameEntityId) { n.id = id }
func (n *NpcEntity) Shape() *cp.Shape      { return n.shape }

// Calculate velocity based on simple pathfinding algorithm between waypoints
func (n *NpcEntity) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// No movement if no active wayPoint or movement paused
	if n.currentWpIndex == -1 || n.stopMovementUntil.After(time.Now()) {
		body.SetVelocityVector(cp.Vector{})
		n.animation = calculateWalkingAnimation(body.Velocity(), n.orientation)
		return
	}
	destination := n.wayPoints[n.currentWpIndex]
	position := body.Position()
	diff := destination.Sub(position)
	diffNormalized := diff.Normalize()

	// Go to next waypoint if in close proximity to current WP
	if diff.Length() < 5 {
		n.currentWpIndex++
		if n.currentWpIndex > len(n.wayPoints)-1 {
			if n.loopWaypoints {
				// Loop back to first index
				n.currentWpIndex = 0
			} else {
				// Quit loop
				n.currentWpIndex = -1
			}
		}
	}
	vel := diffNormalized.Mult(n.velocity)
	body.SetVelocityVector(vel)
	// Update active animation & orientation
	n.orientation = calculateOrientation(vel)
	n.animation = calculateWalkingAnimation(vel, n.orientation)
}
