package loot

type GoldLootTable struct {
	Gold int64
}

type GoldItem struct{ Amount int64 }

func (t *GoldItem) Always() bool         { return true }
func (t *GoldItem) Enabled() bool        { return true }
func (t *GoldItem) Probability() float64 { return 1.0 }
func (t *GoldItem) Value() int64         { return t.Amount }

func (t *GoldLootTable) Always() bool              { return true }
func (t *GoldLootTable) Contents() []LootTableItem { return []LootTableItem{} }
func (t *GoldLootTable) Enabled() bool             { return true }
func (t *GoldLootTable) Probability() float64      { return 1.0 }
func (t *GoldLootTable) Count() int64              { return 1 }
func (t *GoldLootTable) Result() []LootTableItem {
	return []LootTableItem{&GoldItem{t.Gold}}
}

// Shortcut to init a loot table with only gold inside
// TODO: Implement min, max
func NewGoldLootTable(amount int64) LootTable {
	t := &GoldLootTable{Gold: amount}
	return t
}
