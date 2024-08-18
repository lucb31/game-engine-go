package engine

type LootTable struct {
	Gold int64
}

func EmptyLootTable() *LootTable {
	return &LootTable{}
}
