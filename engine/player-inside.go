package engine

import (
	"log"

	"github.com/jakecoffman/cp"
)

type GameEntityEnterable interface {
	GameEntity
	Enter(GameEntityEntering) error
	Leave(GameEntityEntering) error
}

type GameEntityEntering interface {
	GameEntity
	Enter(GameEntityEnterable) error
	Leave() error
}

// Handles all interactions with buildings
type BuildingInteractionController struct {
	// Reference to currently active building
	insideBuilding GameEntityEnterable

	// Reference to entity
	controlledEntity GameEntityEntering
}

func NewBuildingInteractionController(e GameEntityEntering) (*BuildingInteractionController, error) {
	c := &BuildingInteractionController{controlledEntity: e}
	return c, nil
}

func (c *BuildingInteractionController) Inside() bool { return c.insideBuilding != nil }

// Queries physics space for buildings that can be entered
func (c *BuildingInteractionController) BuildingInRange() GameEntityEnterable {
	queryInfo := c.controlledEntity.Shape().Space().PointQueryNearest(c.controlledEntity.Shape().Body().Position(), playerPickupRange, cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, TowerCategory))
	if queryInfo.Shape != nil {
		item, ok := queryInfo.Shape.Body().UserData.(GameEntityEnterable)
		if !ok {
			log.Println("Error: Expected building that can be entered, but received sth else")
		}
		return item
	}
	return nil
}

func (c *BuildingInteractionController) Leave() error {
	log.Println("Controller exiting building...")
	if err := c.insideBuilding.Leave(c.controlledEntity); err != nil {
		return err
	}
	c.insideBuilding = nil
	return nil
}

func (c *BuildingInteractionController) Enter(building GameEntityEnterable) error {
	log.Println("Controller entering building")
	// Check for buildings to enter
	if err := building.Enter(c.controlledEntity); err != nil {
		return err
	}
	c.insideBuilding = building
	// Stop all movement, set position
	c.controlledEntity.Shape().Body().SetVelocity(0, 0)
	c.controlledEntity.Shape().Body().SetPosition(building.Shape().Body().Position())
	return nil
}
