package engine

import (
	"log"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type NpcEntity struct {
	*BaseEntityImpl

	// Logic
	GameEntityStats
	loot loot.LootTable

	// Rendering
	asset       *CharacterAsset
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
	// TODO: Deprecate Waypoints
	WaypointInfo
}

func NpcCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, NpcCategory, PlayerCategory|OuterWallsCategory|HarvestableCategory|TowerCategory|ProjectileCategory)
}

func NewNpc(remover EntityRemover, asset *CharacterAsset, opts NpcOpts) (*NpcEntity, error) {
	base, err := NewBaseEntity(remover)
	if err != nil {
		return nil, err
	}
	npc := &NpcEntity{BaseEntityImpl: base}
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

	// Logic
	npc.armor = 0.0
	npc.maxHealth = 100.0
	npc.movementSpeed = 75.0
	npc.power = 20.0
	npc.loot = loot.NewEmptyLootTable()

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
		npc.loot = loot.NewGoldLootTable(opts.GoldValue)
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
	return n.asset.Draw(t, n.shape, n.orientation)
}

func (n *NpcEntity) Shape() *cp.Shape          { return n.shape }
func (n *NpcEntity) IsVulnerable() bool        { return true }
func (n *NpcEntity) LootTable() loot.LootTable { return n.loot }

func (n *NpcEntity) defaultMovementAI(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	n.simpleWaypointAlgorithm(body, dt)
}

// Calculate velocity based on simple pathfinding algorithm between waypoints
func (n *NpcEntity) simpleWaypointAlgorithm(body *cp.Body, dt float64) {
	// No movement if no active wayPoint
	if n.currentWpIndex == -1 || n.currentWpIndex > len(n.wayPoints)-1 {
		body.SetVelocityVector(cp.Vector{})
		if err := n.asset.AnimationController().Loop("idle"); err != nil {
			log.Fatalln("Error animating npc", err.Error())
			return
		}
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
	if err := n.asset.AnimationController().Loop("walk"); err != nil {
		log.Println("error looping", err.Error())
	}
	return diff.Length()
}

// DEBUG: Draw connecting lines between npcs & waypoints
// for _, wp := range n.wayPoints {
// 	body := cp.NewKinematicBody()
// 	body.SetPosition(wp)
// 	bb := cp.NewBBForCircle(wp, 4)
// 	if err := DrawRectBoundingBox(t, bb); err != nil {
// 		log.Println("err", err.Error())
// 	}
// 	// Draw wp index (top left corner, not centered)
// 	// relWpPos := n.w.camera.AbsToRel(wp)
// 	// ebitenutil.DebugPrintAt(n.world.camera.Screen(), fmt.Sprintf("%d", idx), int(relWpPos.X), int(relWpPos.Y))

// }
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
