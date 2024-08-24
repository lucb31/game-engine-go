package engine

import (
	"github.com/jakecoffman/cp"
)

type TreeEntity struct {
	// Dependencies
	em EntityRemover

	id    GameEntityId
	loot  *LootTable
	shape *cp.Shape
}

const (
	treeWidth  = 16
	treeHeight = 32
)

func NewTree(em EntityRemover, pos cp.Vector) (*TreeEntity, error) {
	t := &TreeEntity{}
	loot := &LootTable{Gold: 10}

	// Init body
	treeBody := cp.NewKinematicBody()
	treeBody.SetPosition(pos)
	treeBody.UserData = t

	// Collision model
	shape := cp.NewBox(treeBody, treeWidth, treeHeight, 0)
	shape.SetElasticity(0)
	shape.SetFriction(0)
	shape.SetFilter(HarvestableCollisionFilter)
	t.shape = shape
	t.loot = loot
	t.em = em

	return t, nil
}

// TODO: Currently only draws bounding box
func (p *TreeEntity) Draw(t RenderingTarget) error {
	DrawRectBoundingBox(t, p.shape)
	return nil
}

var HarvestableCollisionFilter = cp.NewShapeFilter(0, HarvestableCategory, PlayerCategory)

func (p *TreeEntity) Id() GameEntityId      { return p.id }
func (p *TreeEntity) SetId(id GameEntityId) { p.id = id }
func (p *TreeEntity) Shape() *cp.Shape      { return p.shape }
func (p *TreeEntity) LootTable() *LootTable { return p.loot }

func (p *TreeEntity) Destroy() error {
	return p.em.RemoveEntity(p)
}
