package loot

type EmptyLootTable struct{}

func (t *EmptyLootTable) Always() bool              { return false }
func (t *EmptyLootTable) Contents() []LootTableItem { return []LootTableItem{} }
func (t *EmptyLootTable) Enabled() bool             { return true }
func (t *EmptyLootTable) Probability() float64      { return 1.0 }
func (t *EmptyLootTable) Count() int64              { return 1 }
func (t *EmptyLootTable) Result() []LootTableItem   { return []LootTableItem{} }

// Constructor for an empty loot table
func NewEmptyLootTable() *EmptyLootTable {
	return &EmptyLootTable{}
}
