package loot

type ResourceItem struct {
	Amount int64
}

type GoldItem struct {
	*ResourceItem
}

func NewGoldItem(amount int64) *GoldItem {
	resource := &ResourceItem{amount}
	return &GoldItem{ResourceItem: resource}
}

type WoodItem struct {
	*ResourceItem
}

func NewWoodItem(amount int64) *WoodItem {
	resource := &ResourceItem{amount}
	return &WoodItem{ResourceItem: resource}
}

func (t *ResourceItem) Always() bool         { return true }
func (t *ResourceItem) Enabled() bool        { return true }
func (t *ResourceItem) Probability() float64 { return 1.0 }
func (t *ResourceItem) Value() int64         { return t.Amount }

type ResourcesLootTable struct {
	Gold int64
	Wood int64
}

func (t *ResourcesLootTable) Result() []LootTableItem {
	res := []LootTableItem{}
	if t.Gold > 0 {
		res = append(res, NewGoldItem(t.Gold))
	}
	if t.Wood > 0 {
		res = append(res, NewWoodItem(t.Wood))
	}
	return res
}

func (t *ResourcesLootTable) AddWood(amount int64) { t.Wood += amount }
func (t *ResourcesLootTable) AddGold(amount int64) { t.Gold += amount }

func (t *ResourcesLootTable) Always() bool              { return true }
func (t *ResourcesLootTable) Contents() []LootTableItem { return []LootTableItem{} }
func (t *ResourcesLootTable) Enabled() bool             { return true }
func (t *ResourcesLootTable) Probability() float64      { return 1.0 }
func (t *ResourcesLootTable) Count() int64              { return 1 }

func NewResourcesLootTable() *ResourcesLootTable {
	return &ResourcesLootTable{}
}

// Shortcut to init a loot table with only gold inside
// TODO: Implement min, max
func NewGoldLootTable(amount int64) *ResourcesLootTable {
	t := NewResourcesLootTable()
	t.AddGold(amount)
	return t
}
