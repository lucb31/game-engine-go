package loot

type ItemEffectId = int64

const (
	ItemEffectAddMaxHealth ItemEffectId = iota
	ItemEffectAddMovementSpeed
	ItemEffectAddArmor
	ItemEffectAddPower
	ItemEffectAddAtkSpeed
	ItemEffectAddCastleProjectile
)

// Item definition stored in ItemDB, not copy / instance of one particular item
type GameItem struct {
	// Store
	GoldPrice   int64
	WoodPrice   int64
	Description string

	// Effect
	ItemEffectId

	// Inventory
	StackSize int64
}
