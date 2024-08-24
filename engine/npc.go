package engine

import (
	"github.com/jakecoffman/cp"
)

type NpcEntity struct {
	id      GameEntityId
	remover EntityRemover

	// Logic
	GameEntityStats
	loot *LootTable

	// Rendering
	asset       *CharacterAsset
	animation   string
	orientation Orientation

	// Physics
	shape *cp.Shape

	// Movement AI
	wayPoints      []cp.Vector
	currentWpIndex int
	loopWaypoints  bool
}

type NpcOpts struct {
	StartingPos       cp.Vector
	BaseArmor         float64
	BasePower         float64
	BaseHealth        float64
	BaseMovementSpeed float64
	GoldValue         int64
	Waypoints         []cp.Vector
}

func NpcCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, NpcCategory, PlayerCategory|OuterWallsCategory|HarvestableCategory|TowerCategory|ProjectileCategory)
}

func NewNpc(remover EntityRemover, asset *CharacterAsset, opts NpcOpts) (*NpcEntity, error) {
	npc := &NpcEntity{remover: remover}
	// Physics model
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 48, Y: 16})
	body.SetVelocityUpdateFunc(npc.defaultMovementAI)
	body.UserData = npc

	// Collision model
	npc.shape = cp.NewBox(body, 32, 32, 0)
	npc.shape.SetElasticity(0)
	npc.shape.SetFriction(1)
	npc.shape.SetCollisionType(cp.CollisionType(NpcCollision))
	npc.shape.SetFilter(NpcCollisionFilter())

	// Asset
	npc.asset = asset
	npc.animation = "idle_east"

	// Logic
	npc.armor = 0.0
	npc.maxHealth = 100.0
	npc.movementSpeed = 75.0
	npc.power = 20.0
	npc.loot = EmptyLootTable()

	// AI
	npc.loopWaypoints = false

	// Parse opts
	if opts.BaseArmor > 0 {
		npc.armor = opts.BaseArmor
	}
	if opts.BaseHealth > 0 {
		npc.maxHealth = opts.BaseHealth
	}
	if opts.BaseMovementSpeed > 0 {
		npc.movementSpeed = opts.BaseMovementSpeed
	}
	if opts.BasePower > 0 {
		npc.power = opts.BasePower
	}
	if opts.GoldValue > 0 {
		npc.loot.Gold = opts.GoldValue
	}
	if opts.StartingPos.Length() > 0 {
		body.SetPosition(opts.StartingPos)
	}
	if len(opts.Waypoints) > 0 {
		npc.wayPoints = opts.Waypoints
	}

	npc.health = npc.maxHealth
	return npc, nil
}

func (n *NpcEntity) Draw(t RenderingTarget) error {
	n.asset.DrawHealthbar(t, n.shape, n.health, n.maxHealth)
	return n.asset.Draw(t, n.animation, n.shape)
}

func (n *NpcEntity) Destroy() error {
	return n.remover.RemoveEntity(n)
}

func (n *NpcEntity) Id() GameEntityId      { return n.id }
func (n *NpcEntity) SetId(id GameEntityId) { n.id = id }
func (n *NpcEntity) Shape() *cp.Shape      { return n.shape }
func (n *NpcEntity) IsVulnerable() bool    { return true }
func (n *NpcEntity) LootTable() *LootTable { return n.loot }

func (n *NpcEntity) defaultMovementAI(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	n.simpleWaypointAlgorithm(body, dt)
}

// Calculate velocity based on simple pathfinding algorithm between waypoints
func (n *NpcEntity) simpleWaypointAlgorithm(body *cp.Body, dt float64) {
	// No movement if no active wayPoint
	if n.currentWpIndex == -1 || n.currentWpIndex > len(n.wayPoints)-1 {
		body.SetVelocityVector(cp.Vector{})
		n.animation = calculateWalkingAnimation(body.Velocity(), n.orientation)
		return
	}
	destination := n.wayPoints[n.currentWpIndex]
	distance := n.moveTowards(body, destination)

	// Go to next waypoint if in close proximity to current WP
	// ~Distance covered within next timestep
	dx := n.movementSpeed * dt
	if distance < dx {
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

}

// Updates body velocity to move towards destination
// Returns remaining distance
func (n *NpcEntity) moveTowards(body *cp.Body, dest cp.Vector) float64 {
	position := body.Position()
	diff := dest.Sub(position)
	diffNormalized := diff.Normalize()

	vel := diffNormalized.Mult(n.movementSpeed)
	body.SetVelocityVector(vel)
	// Update active animation & orientation
	n.orientation = updateOrientation(n.orientation, vel)
	n.animation = calculateWalkingAnimation(vel, n.orientation)
	return diff.Length()
}
