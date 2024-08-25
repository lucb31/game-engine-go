package loot

type Inventory interface {
	Add(LootTable) error
	CanAfford(int64) bool
	Balance() int64
	Revenue() int64
	Spend(int64) (int64, error)
}

type InMemoryInventory struct {
	goldManager GoldManager
}

func NewInventory() (*InMemoryInventory, error) {
	goldManager, err := NewInMemoryGoldManager()
	if err != nil {
		return nil, err
	}
	// Add starting gold
	goldManager.Add(5000)

	inv := &InMemoryInventory{goldManager: goldManager}
	return inv, nil
}

func (i *InMemoryInventory) Add(loot LootTable) error {
	lootResult := loot.Result()
	// FIX: Currently only supports gold drops. Ignores all other
	for _, lootItem := range lootResult {
		goldItem, ok := lootItem.(*GoldItem)
		if ok {
			_, err := i.goldManager.Add(goldItem.Value())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *InMemoryInventory) Balance() int64 {
	return i.goldManager.Balance()
}

func (i *InMemoryInventory) Revenue() int64 {
	return i.goldManager.Revenue()
}

func (i *InMemoryInventory) CanAfford(amount int64) bool {
	return i.goldManager.CanAfford(amount)
}

func (i *InMemoryInventory) Spend(amount int64) (int64, error) {
	return i.goldManager.Remove(amount)
}
