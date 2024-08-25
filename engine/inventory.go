package engine

type Inventory interface {
	Add(*LootTable) error
	CanAfford(int64) bool
	Balance() int64
	Revenue() int64
	Spend(int64) (int64, error)
}

type InMemoryInventory struct {
	goldManager GoldManager
}

type GameEntityWithInventory interface {
	GameEntity
	Inventory() Inventory
}

func NewInventory() (*InMemoryInventory, error) {
	goldManager, err := NewInMemoryGoldManager()
	if err != nil {
		return nil, err
	}
	// Add starting gold
	goldManager.Add(50)

	inv := &InMemoryInventory{goldManager: goldManager}
	return inv, nil
}

func (i *InMemoryInventory) Add(loot *LootTable) error {
	_, err := i.goldManager.Add(loot.Gold)
	if err != nil {
		return err
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
