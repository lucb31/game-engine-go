package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type HarvestingTool interface {
	// Returns true if there is a harvestable entity within range
	InRange() bool
	// Returns nil if nothing in range
	Nearest() Harvestable

	// Start harvesting nearest harvestable object
	HarvestNearest() error
	Abort() error
}

type Harvestable interface {
	BaseEntity
	Lootable
}

var HarvestToolCollisionFilter = cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, HarvestableCategory)
var HarvestableCollisionFilter = cp.NewShapeFilter(0, HarvestableCategory, PlayerCategory)

const (
	defaultHarvestingSpeed = 1.0
	defaultHarvestingRange = 50.0
)

type WoodHarvestingTool struct {
	// Dependencies
	owner GameEntityWithInventory

	// Stats
	harvestingRange float64
	harvestingSpeed float64

	// State
	harvestingTimer *IngameTimer
	target          Harvestable
}

func NewWoodHarvestingTool(itp IngameTimeProvider, owner GameEntityWithInventory) (*WoodHarvestingTool, error) {
	if owner == nil {
		return nil, fmt.Errorf("Cannot init wood harvesting tool without owner")
	}
	ht := &WoodHarvestingTool{owner: owner}
	// Setup timer
	var err error
	if ht.harvestingTimer, err = NewIngameTimer(itp); err != nil {
		return nil, err
	}
	// Set defaults
	ht.harvestingRange = defaultHarvestingRange
	ht.harvestingSpeed = defaultHarvestingSpeed
	return ht, nil
}

func (ht *WoodHarvestingTool) Nearest() Harvestable {
	query := ht.owner.Shape().Space().PointQueryNearest(ht.owner.Shape().Body().Position(), ht.harvestingRange, HarvestToolCollisionFilter)
	if query.Shape == nil {
		return nil
	}
	harvestable, ok := query.Shape.Body().UserData.(Harvestable)
	if !ok {
		fmt.Println("Expected harvestable target, but found something else", query.Shape.Body().UserData)
		return nil
	}
	return harvestable
}

func (ht *WoodHarvestingTool) InRange() bool {
	return ht.Nearest() != nil
}

func (ht *WoodHarvestingTool) HarvestNearest() error {
	// Initiate if not already harvesting
	if !ht.harvestingTimer.Active() {
		target := ht.Nearest()
		if target == nil {
			return fmt.Errorf("Nothin in range")
		}
		fmt.Println("Starting harvest", target)
		ht.target = target
		ht.harvestingTimer.Start()
		return nil
	}

	// Check if done
	if ht.harvestingTimer.Elapsed() > ht.harvestingSpeed {
		fmt.Println("We're done!")
		if err := ht.target.Destroy(); err != nil {
			return err
		}
		// Add loot for harvesting target
		if err := ht.owner.Inventory().AddLoot(ht.target.LootTable()); err != nil {
			return err
		}

		// Reset harvesting tool
		ht.harvestingTimer.Stop()
		ht.target = nil
	}
	return nil
}

func (ht *WoodHarvestingTool) Abort() error {
	if ht.harvestingTimer.Active() {
		fmt.Println("aborting harvest")
		ht.harvestingTimer.Stop()
	}
	return nil
}
