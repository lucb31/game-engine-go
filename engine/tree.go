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

	// Visuals
	asset *CharacterAsset
}

const (
	treeWidth  = 32
	treeHeight = 64
)

func NewTree(em EntityRemover, pos cp.Vector, asset *CharacterAsset) (*TreeEntity, error) {
	base, err := NewBaseEntity(em)
	if err != nil {
		return nil, err
	}
	t := &TreeEntity{BaseEntityImpl: base, asset: asset}
	loot := loot.NewResourcesLootTable()
	loot.AddWood(5)

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

func (p *TreeEntity) Draw(t RenderingTarget) error {
	p.asset.Draw(t, p.shape, 0)
	return nil
}

func (p *TreeEntity) Id() GameEntityId          { return p.id }
func (p *TreeEntity) SetId(id GameEntityId)     { p.id = id }
func (p *TreeEntity) Shape() *cp.Shape          { return p.shape }
func (p *TreeEntity) LootTable() loot.LootTable { return p.loot }
