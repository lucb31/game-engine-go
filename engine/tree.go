package engine

import (
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type TreeEntity struct {
	// Dependencies
	*BaseEntityImpl
	loot  loot.LootTable
	shape *cp.Shape
}

const (
	treeWidth  = 32
	treeHeight = 64
)

func NewTree(em EntityRemover, pos cp.Vector) (*TreeEntity, error) {
	base, err := NewBaseEntity(em)
	if err != nil {
		return nil, err
	}
	t := &TreeEntity{BaseEntityImpl: base}
	loot := loot.NewGoldLootTable(10)

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

	return t, nil
}

// TODO: Currently only draws bounding box
func (p *TreeEntity) Draw(t RenderingTarget) error {
	DrawRectBoundingBox(t, p.shape)
	return nil
}

var HarvestableCollisionFilter = cp.NewShapeFilter(0, HarvestableCategory, PlayerCategory)

func (p *TreeEntity) Id() GameEntityId          { return p.id }
func (p *TreeEntity) SetId(id GameEntityId)     { p.id = id }
func (p *TreeEntity) Shape() *cp.Shape          { return p.shape }
func (p *TreeEntity) LootTable() loot.LootTable { return p.loot }
