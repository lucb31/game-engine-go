package td

import (
	"fmt"

	"github.com/lucb31/game-engine-go/engine"
)

type CreepManager struct {
	entityManager           engine.GameEntityManager
	goldManger              engine.GoldManager
	asset                   *engine.CharacterAsset
	creepsSpawned           int
	lastCreepSpawnedTime    float64
	creepSpawnRatePerSecond float64
	creepsToSpawn           int
}

func NewCreepManager(em engine.GameEntityManager, asset *engine.CharacterAsset, goldManager engine.GoldManager) (*CreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	return &CreepManager{entityManager: em, asset: asset, creepSpawnRatePerSecond: 0.5, creepsToSpawn: 30, goldManger: goldManager}, nil
}

func (c *CreepManager) Update() error {
	return c.spawnCreep()
}

func (c *CreepManager) spawnCreep() error {
	// All creeps spawned
	if c.creepsSpawned >= c.creepsToSpawn {
		return nil
	}
	// Timeout until creep spawn timer over
	now := c.entityManager.GetIngameTime()
	diff := now - c.lastCreepSpawnedTime
	if diff < 1/c.creepSpawnRatePerSecond {
		return nil
	}
	// Initialize an npc
	npc, err := engine.NewNpc(c.entityManager, c.asset, c.goldManger)
	if err != nil {
		return err
	}
	c.entityManager.AddEntity(npc)
	c.lastCreepSpawnedTime = now
	c.creepsSpawned++
	return nil
}

func (c *CreepManager) GetProgress() ProgressInfo {
	return ProgressInfo{min: 0, max: c.creepsToSpawn, current: c.creepsSpawned}
}
