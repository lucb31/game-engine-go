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
