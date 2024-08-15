package td

import (
	"fmt"

	"github.com/lucb31/game-engine-go/engine"
)

type CreepManager struct {
	entityManager engine.GameEntityManager
	goldManager   engine.GoldManager
	asset         *engine.CharacterAsset

	activeWave int

	// Active wave
	creepsSpawned        int
	creepsAlive          int
	lastCreepSpawnedTime float64
}

const goldPerKill = int64(1)

func NewCreepManager(em engine.GameEntityManager, asset *engine.CharacterAsset, goldManager engine.GoldManager) (*CreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	return &CreepManager{entityManager: em, asset: asset, activeWave: 1, goldManager: goldManager}, nil
}

func (c *CreepManager) Update() error {
	// Check if creeps to spawn left
	if c.creepsSpawned < c.creepsToSpawn() {
		if err := c.spawnCreep(); err != nil {
			return err
		}
	} else if c.creepsAlive == 0 {
		// Wave cleared
		return c.nextWave()
	}
	return nil
}

func (c *CreepManager) spawnCreep() error {
	// Timeout until creep spawn timer over
	now := c.entityManager.GetIngameTime()
	diff := now - c.lastCreepSpawnedTime
	if diff < 1/c.creepSpawnRatePerSecond() {
		return nil
	}
	// Initialize an npc
	npc, err := engine.NewNpc(c, c.asset, c.goldManager)
	if err != nil {
		return err
	}
	c.entityManager.AddEntity(npc)
	c.lastCreepSpawnedTime = now
	c.creepsAlive++
	c.creepsSpawned++
	return nil
}

func (c *CreepManager) nextWave() error {
	c.activeWave++
	c.creepsAlive = 0
	c.creepsSpawned = 0
	fmt.Printf("Wave cleared! Starting wave %d...\n", c.activeWave)
	return nil
}

func (c *CreepManager) RemoveEntity(entity engine.GameEntity) error {
	// Remove npc from game world
	if err := c.entityManager.RemoveEntity(entity); err != nil {
		return err
	}
	c.creepsAlive--
	// Add gold for kill
	_, err := c.goldManager.Add(goldPerKill)
	return err
}

func (c *CreepManager) creepSpawnRatePerSecond() float64 {
	return 0.5
}

func (c *CreepManager) creepsToSpawn() int {
	return 30
}

func (c *CreepManager) GetProgress() ProgressInfo {
	label := fmt.Sprintf("Wave %d", c.activeWave)
	return ProgressInfo{min: 0, max: c.creepsToSpawn(), current: c.creepsSpawned, label: label}
}
