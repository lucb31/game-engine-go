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
	// Remaining health until fully harvested
	health float64
	// Maximum health to harvest
	maxHealth float64

	// Visuals
	asset *CharacterAsset
}

func newTreeEntity(asset *CharacterAsset, width, height float64) (*TreeEntity, error) {
	base, err := NewBaseEntity()
	if err != nil {
		return nil, err
	}
	t := &TreeEntity{BaseEntityImpl: base, asset: asset}

	// Init body
	treeBody := cp.NewKinematicBody()
	treeBody.UserData = t

	// Collision model
	shape := cp.NewBox(treeBody, width, height, 0)
	shape.SetElasticity(0)
	shape.SetFriction(0)
	shape.SetFilter(HarvestableCollisionFilter)
	t.shape = shape
	t.health = 100.0
	t.maxHealth = 100.0

	return t, nil
}

func NewTree(asset *CharacterAsset) (*TreeEntity, error) {
	t, err := newTreeEntity(asset, 32.0, 64.0)
	if err != nil {
		return nil, err
	}
	loot := loot.NewResourcesLootTable()
	loot.AddWood(5)
	t.loot = loot
	return t, nil
}

func NewBush(asset *CharacterAsset) (*TreeEntity, error) {
	t, err := newTreeEntity(asset, 32.0, 32.0)
	if err != nil {
		return nil, err
	}
	loot := loot.NewResourcesLootTable()
	loot.AddWood(2)
	t.loot = loot
	t.health = 50.0
	t.maxHealth = 50.0
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
func (p *TreeEntity) Position() cp.Vector       { return p.Shape().Body().Position() }
func (p *TreeEntity) Health() float64           { return p.health }
func (p *TreeEntity) SetHealth(v float64)       { p.health = v }
func (p *TreeEntity) Armor() float64            { return 0 }
func (p *TreeEntity) IsVulnerable() bool        { return true }
func (p *TreeEntity) SetPosition(pos cp.Vector) { p.Shape().Body().SetPosition(pos) }
