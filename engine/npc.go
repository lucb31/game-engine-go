package engine

import (
	"github.com/jakecoffman/cp"
)

type NpcEntity struct {
	id      GameEntityId
	remover EntityRemover

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
	wayPoints      []cp.Vector
	currentWpIndex int
	loopWaypoints  bool
}

type NpcOpts struct {
	StartingHealth float64
	StartingPos    cp.Vector
}

func NpcCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, NpcCategory, PlayerCategory|OuterWallsCategory|InnerWallsCategory|TowerCategory|ProjectileCategory)
}

func NewNpc(remover EntityRemover, asset *CharacterAsset, opts NpcOpts) (*NpcEntity, error) {
	npc := &NpcEntity{remover: remover, orientation: South}
	// Physics model
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 48, Y: 16})
	body.SetVelocityUpdateFunc(npc.calculateVelocity)
	body.UserData = npc

	// Collision model
	npc.shape = cp.NewBox(body, 32, 32, 0)
	npc.shape.SetElasticity(0)
	npc.shape.SetFriction(1)
	npc.shape.SetCollisionType(cp.CollisionType(NpcCollision))
	npc.shape.SetFilter(NpcCollisionFilter())

	npc.asset = asset
	npc.wayPoints = []cp.Vector{
		{X: 48, Y: 720},
		{X: 976, Y: 720},
		{X: 976, Y: 48},
		{X: 208, Y: 48},
		{X: 208, Y: 560},
		{X: 816, Y: 560},
		{X: 816, Y: 208},
		{X: 368, Y: 208},
		{X: 368, Y: 384},
		{X: 640, Y: 384},
	}
	npc.loopWaypoints = false
	npc.velocity = 75.0
	npc.health = 100.0
	npc.animation = "idle_south"

	// Parse opts
	if opts.StartingHealth > 0 {
		npc.health = opts.StartingHealth
	}
	if opts.StartingPos.Length() > 0 {
		body.SetPosition(opts.StartingPos)
	}

	return npc, nil
}

func (n *NpcEntity) Draw(t RenderingTarget) {
	n.asset.Draw(t, n.animation, n.shape)
}

func (n *NpcEntity) Destroy() error {
	return n.remover.RemoveEntity(n)
}

func (n *NpcEntity) Id() GameEntityId      { return n.id }
func (n *NpcEntity) SetId(id GameEntityId) { n.id = id }
func (n *NpcEntity) Shape() *cp.Shape      { return n.shape }
func (n *NpcEntity) Armor() float64        { return 0.0 }
func (n *NpcEntity) Health() float64       { return n.health }
func (n *NpcEntity) SetHealth(h float64)   { n.health = h }

// Calculate velocity based on simple pathfinding algorithm between waypoints
func (n *NpcEntity) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// No movement if no active wayPoint
	if n.currentWpIndex == -1 {
		body.SetVelocityVector(cp.Vector{})
		n.animation = calculateWalkingAnimation(body.Velocity(), n.orientation)
		return
	}
	destination := n.wayPoints[n.currentWpIndex]
	position := body.Position()
	diff := destination.Sub(position)
	diffNormalized := diff.Normalize()

	// Go to next waypoint if in close proximity to current WP
	// ~Distance covered within next timestep
	dx := n.velocity * dt
	if diff.Length() < dx {
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
