package engine

import (
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/loot"
)

// Entity representing a sprite that can be picked up by the player
type ItemEntity struct {
	*BaseEntityImpl
	shape *cp.Shape
	asset *CharacterAsset
	loot  loot.LootTable
}

var itemCollisionFilter = cp.NewShapeFilter(cp.NO_GROUP, ItemCategory, PlayerCategory)

func NewItemEntity(pos cp.Vector) (*ItemEntity, error) {
	base, err := NewBaseEntity()
	if err != nil {
		return nil, err
	}
	item := &ItemEntity{BaseEntityImpl: base}

	body := cp.NewKinematicBody()
	body.SetPosition(pos)
	body.UserData = item

	item.shape = cp.NewCircle(body, 8, cp.Vector{})
	item.shape.SetFilter(itemCollisionFilter)
	item.shape.SetCollisionType(ItemCollision)

	item.loot = loot.NewEmptyLootTable()

	return item, nil
}

func (i *ItemEntity) Draw(t RenderingTarget) error {
	if i.asset != nil {
		if err := i.asset.Draw(t, i.shape, 0); err != nil {
			return err
		}
	} else {
		if err := DrawRectBoundingBox(t, i.shape.BB()); err != nil {
			return err
		}
	}
	return nil
}

func (i *ItemEntity) SetLootTable(loot loot.LootTable) { i.loot = loot }
func (i *ItemEntity) SetAsset(asset *CharacterAsset) error {
	i.asset = asset
	if err := i.asset.AnimationController().Loop("idle"); err != nil {
		return err
	}
	if err := i.asset.AnimationController().Play("spawn"); err != nil {
		return err
	}
	return nil
}
func (i *ItemEntity) LootTable() loot.LootTable { return i.loot }
func (i *ItemEntity) Shape() *cp.Shape          { return i.shape }
