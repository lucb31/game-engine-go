package loot

type ItemEffectId = int64

const (
	ItemEffectAddMaxHealth ItemEffectId = iota
	ItemEffectAddMovementSpeed
	ItemEffectAddArmor
	ItemEffectAddPower
	ItemEffectAddAtkSpeed
)

// Item definition, not instance stored in ItemDB
type GameItem struct {
	// Store
	Price       int64
	Description string

	// Effect
	ItemEffectId

	// Inventory
	StackSize int64
}
