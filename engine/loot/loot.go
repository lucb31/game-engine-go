package loot

// Inspired by https://www.codeproject.com/Articles/420046/Loot-Tables-Random-Maps-and-Monsters-Part-I
type LootTableItem interface {
	Probability() float64
	Always() bool
	Enabled() bool
}

type LootTableItemValue interface {
	LootTableItem
	Value() int64
}

type LootTable interface {
	LootTableItem

	// Used for recursion
	Contents() []LootTableItem
	Count() int64
	// Result of (usually randomly) evaluating loot table
	Result() []LootTableItem
}

type GuaranteedLootTable struct {
	items []LootTableItem
}

func (t *GuaranteedLootTable) Always() bool              { return false }
func (t *GuaranteedLootTable) Contents() []LootTableItem { return t.items }
func (t *GuaranteedLootTable) Enabled() bool             { return true }
func (t *GuaranteedLootTable) Probability() float64      { return 1.0 }
func (t *GuaranteedLootTable) Count() int64              { return 1 }
func (t *GuaranteedLootTable) Result() []LootTableItem   { return t.items }

// Constructor for guaranteed drop loot table
func NewGuaranteedLootTable(item LootTableItem) *GuaranteedLootTable {
	return &GuaranteedLootTable{[]LootTableItem{item}}
}
