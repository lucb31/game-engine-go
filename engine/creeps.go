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
	SetProvider(c CreepProvider) error
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
	Round              int
	TotalCreepsToSpawn int
	WaveTicksPerSecond float64
	CreepsPerTick      int
	HealthScalingFunc  func(baseHealth float64) float64
}

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
	return cm, nil
}

func (c *BaseCreepManager) Update() error {
	// Check if creeps to spawn left
	if c.creepsSpawned < c.activeWave.TotalCreepsToSpawn {
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
	loot := *entity.LootTable()
	// Remove npc from game world
	if err := c.entityManager.RemoveEntity(entity); err != nil {
		return err
	}
	c.creepsAlive--
	// Add gold for kill
	// TODO: Should be handled via event-based damage / loot system
	_, err := c.goldManager.Add(loot.Gold)
	return err
}

func (c *BaseCreepManager) Progress() hud.ProgressInfo {
	label := fmt.Sprintf("Wave %d", c.activeWave.Round)
	return hud.ProgressInfo{Min: 0, Max: c.activeWave.TotalCreepsToSpawn, Current: c.creepsSpawned, Label: label}
}

func (c *BaseCreepManager) Round() int { return c.activeWave.Round }
func (c *BaseCreepManager) SetProvider(p CreepProvider) error {
	c.creepProvider = p
	return c.NextWave()
}

func (c *BaseCreepManager) spawnCreep() error {
	// Timeout until creep spawn timer over
	now := c.entityManager.GetIngameTime()
	diff := now - c.lastCreepSpawnedTime
	if diff < 1/c.activeWave.WaveTicksPerSecond {
		return nil
	}

	// Initialize an npc
	for i := 0; i < c.activeWave.CreepsPerTick && c.creepsSpawned < c.activeWave.TotalCreepsToSpawn; i++ {
		npc, err := c.creepProvider.NextNpc(c, *c.activeWave)
		if err != nil {
			return err
		}
		c.entityManager.AddEntity(npc)
		c.creepsAlive++
		c.creepsSpawned++
	}

	// Update metrics
	c.lastCreepSpawnedTime = now
	return nil
}

// Wave scaling
func calculateWaveOpts(round int) Wave {
	wave := Wave{Round: round}
	wave.TotalCreepsToSpawn = int(math.Exp(float64(round)/4) + 29)
	wave.WaveTicksPerSecond = 0.5
	wave.CreepsPerTick = 5
	wave.HealthScalingFunc = func(baseHealth float64) float64 { return math.Pow(3.5*float64(round-1), 2) + baseHealth }
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
