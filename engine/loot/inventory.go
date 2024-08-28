package loot

import (
	"fmt"
	"log"
)

type Inventory interface {
	GoldManager() ResourceManager
	WoodManager() ResourceManager
	AddLoot(LootTable) error
	Buy(*GameItem) error
	CanAfford(*GameItem) (bool, error)
}

type InMemoryInventory struct {
	goldManager ResourceManager
	woodManager ResourceManager
}

func NewInventory() (*InMemoryInventory, error) {
	goldManager, err := NewInMemoryResourceManager()
	if err != nil {
		return nil, err
	}
	// Add starting gold
	goldManager.Add(5000)

	woodManager, err := NewInMemoryResourceManager()
	if err != nil {
		return nil, err
	}

	inv := &InMemoryInventory{goldManager: goldManager, woodManager: woodManager}
	return inv, nil
}

func (i *InMemoryInventory) AddLoot(loot LootTable) error {
	// Evaluate loot table
	lootResult := loot.Result()

	// Add loot table results to inventory
	for _, lootItem := range lootResult {
		log.Println("processing loot item", lootItem)
		switch v := lootItem.(type) {
		case *GoldItem:
			_, err := i.goldManager.Add(v.Value())
			if err != nil {
				return err
			}
		case *WoodItem:
			_, err := i.woodManager.Add(v.Value())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Dont know how to handle loot item: %v", lootItem)
		}
	}
	return nil
}

func (i *InMemoryInventory) Buy(item *GameItem) error {
	// Check funds
	canAfford, err := i.CanAfford(item)
	if !canAfford {
		return fmt.Errorf("Cannot afford item %v: %s", item, err.Error())
	}
	// Update resources
	_, err = i.GoldManager().Remove(item.GoldPrice)
	if err != nil {
		return fmt.Errorf("Error removing item gold cost: %s", err.Error())
	}
	_, err = i.WoodManager().Remove(item.WoodPrice)
	if err != nil {
		return fmt.Errorf("Error removing item wood cost: %s", err.Error())
	}
	return nil
}

func (i *InMemoryInventory) CanAfford(item *GameItem) (bool, error) {
	if !i.goldManager.CanAfford(item.GoldPrice) {
		return false, fmt.Errorf("Insufficient gold funds")
	}
	if !i.woodManager.CanAfford(item.WoodPrice) {
		return false, fmt.Errorf("Insufficient wood funds")
	}
	return true, nil
}

func (i *InMemoryInventory) GoldManager() ResourceManager { return i.goldManager }
func (i *InMemoryInventory) WoodManager() ResourceManager { return i.woodManager }
