package td

import (
	"fmt"
	"time"

	"github.com/lucb31/game-engine-go/engine"
)

type CreepManager struct {
	entityManager           engine.GameEntityManager
	asset                   *engine.CharacterAsset
	creepCount              int
	lastCreepSpawned        time.Time
	creepSpawnRatePerSecond float64
}

func NewCreepManager(em engine.GameEntityManager, asset *engine.CharacterAsset) (*CreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	return &CreepManager{entityManager: em, asset: asset, creepSpawnRatePerSecond: 0.5}, nil
}

func (c *CreepManager) Update() error {
	return c.spawnCreep()
}

func (c *CreepManager) spawnCreep() error {
	now := time.Now()
	duration := float64(time.Second) / c.creepSpawnRatePerSecond
	if now.Sub(c.lastCreepSpawned) < time.Duration(duration) {
		return nil
	}
	// Initialize an npc
	npc, err := engine.NewNpc(c.entityManager, c.asset)
	if err != nil {
		return err
	}
	c.entityManager.AddEntity(npc)
	c.lastCreepSpawned = now
	return nil
}
