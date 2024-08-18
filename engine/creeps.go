package engine

import (
	"fmt"
	"math"

	"github.com/lucb31/game-engine-go/engine/hud"
)

type CreepManager interface {
	Update() error
	Progress() hud.ProgressInfo
	Round() int
}

type BaseCreepManager struct {
	entityManager GameEntityManager
	goldManager   GoldManager
	creepProvider CreepProvider

	activeWave *Wave

	// Active wave
	creepsSpawned        int
	creepsAlive          int
	lastCreepSpawnedTime float64
}

type Wave struct {
	Round                   int
	CreepsToSpawn           int
	CreepOpts               NpcOpts
	CreepSpawnRatePerSecond float64
}

const goldPerKill = int64(2)

func NewBaseCreepManager(em GameEntityManager, goldManager GoldManager) (*BaseCreepManager, error) {
	cm := &BaseCreepManager{entityManager: em, goldManager: goldManager}
	return cm, nil
}

func NewDefaultCreepManager(em GameEntityManager, asset *CharacterAsset, goldManager GoldManager) (*BaseCreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	cm, err := NewBaseCreepManager(em, goldManager)
	if err != nil {
		return nil, err
	}
	creepProvider, err := NewDefaultCreepProvider(asset)
	if err != nil {
		return nil, err
	}
	cm.SetProvider(creepProvider)
	if err = cm.NextWave(); err != nil {
		return nil, err
	}
	return cm, nil
}

func (c *BaseCreepManager) Update() error {
	// Check if creeps to spawn left
	if c.creepsSpawned < c.creepsToSpawn() {
		if err := c.spawnCreep(); err != nil {
			return err
		}
	} else if c.creepsAlive == 0 {
		// Wave cleared
		return c.NextWave()
	}
	return nil
}

func (c *BaseCreepManager) RemoveEntity(entity GameEntity) error {
	// Remove npc from game world
	if err := c.entityManager.RemoveEntity(entity); err != nil {
		return err
	}
	c.creepsAlive--
	// Add gold for kill
	_, err := c.goldManager.Add(goldPerKill)
	return err
}

func (c *BaseCreepManager) Progress() hud.ProgressInfo {
	label := fmt.Sprintf("Wave %d", c.activeWave.Round)
	return hud.ProgressInfo{Min: 0, Max: c.creepsToSpawn(), Current: c.creepsSpawned, Label: label}
}

func (c *BaseCreepManager) Round() int                  { return c.activeWave.Round }
func (c *BaseCreepManager) SetProvider(p CreepProvider) { c.creepProvider = p }

func (c *BaseCreepManager) spawnCreep() error {
	// Timeout until creep spawn timer over
	now := c.entityManager.GetIngameTime()
	diff := now - c.lastCreepSpawnedTime
	if diff < 1/c.creepSpawnRatePerSecond() {
		return nil
	}

	// Initialize an npc
	npc, err := c.creepProvider.NextNpc(c, c.activeWave.CreepOpts)
	if err != nil {
		return err
	}
	c.entityManager.AddEntity(npc)

	// Update metrics
	c.lastCreepSpawnedTime = now
	c.creepsAlive++
	c.creepsSpawned++
	return nil
}

// Wave scaling
func calculateWaveOpts(round int) Wave {
	wave := Wave{Round: round}
	wave.CreepsToSpawn = int(math.Exp(float64(round)/4) + 29)
	wave.CreepSpawnRatePerSecond = 1.0
	startingHealth := math.Pow(3.5*float64(round), 2) + 100
	wave.CreepOpts = NpcOpts{StartingHealth: startingHealth}
	return wave
}

func (c *BaseCreepManager) NextWave() error {
	nextRound := 1
	if c.activeWave != nil {
		nextRound = c.activeWave.Round + 1
	}
	wave := calculateWaveOpts(nextRound)
	c.activeWave = &wave
	c.creepsAlive = 0
	c.creepsSpawned = 0
	fmt.Printf("Wave cleared! Starting wave %v...\n", c.activeWave)
	return nil
}

func (c *BaseCreepManager) creepSpawnRatePerSecond() float64 {
	return c.activeWave.CreepSpawnRatePerSecond
}

func (c *BaseCreepManager) creepsToSpawn() int {
	return c.activeWave.CreepsToSpawn
}
