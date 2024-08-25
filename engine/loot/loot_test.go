package loot

import "testing"

func TestGoldLoot(t *testing.T) {
	goldVal := int64(50)
	lootTable := NewGoldLootTable(goldVal)
	res := lootTable.Result()
	if len(res) == 0 {
		t.Fatalf("Empty loot table")
	}
	goldItem, ok := res[0].(*GoldItem)
	if !ok {
		t.Fatalf("Expected gold item, but could not cast")
	}
	if goldItem.Value() != goldVal {
		t.Fatalf("Expected gold value %d, but received %d", goldVal, goldItem.Value())
	}
}
