package engine

import (
	"fmt"
	"log"
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
	creepProvider CreepProvider

	// Configuration params for currently active wave
	activeWave *Wave

	// Active wave
	creepsSpawned int
	creepsAlive   int
	waveCleared   bool
	// Timer to control creep spawns during an active wave
	creepSpawnTimeout Timeout

	// Timer to control idle time between waves
	spawnIdleTimeout Timeout
}

const idleTimeAfterWaveFinished = 10.0

type Wave struct {
	Round              int
	TotalCreepsToSpawn int
	WaveTicksPerSecond float64
	CreepsPerTick      int
	HealthScalingFunc  func(baseHealth float64) float64
}

func NewBaseCreepManager(em GameEntityManager) (*BaseCreepManager, error) {
	cm := &BaseCreepManager{entityManager: em}
	var err error
	if cm.creepSpawnTimeout, err = NewIngameTimeout(em); err != nil {
		return nil, err
	}
	if cm.spawnIdleTimeout, err = NewIngameTimeout(em); err != nil {
		return nil, err
	}
	// Start with a idle phase
	cm.spawnIdleTimeout.Set(idleTimeAfterWaveFinished)

	return cm, nil
}

func NewDefaultCreepManager(em GameEntityManager, asset *CharacterAsset) (*BaseCreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	cm, err := NewBaseCreepManager(em)
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
	// As long as the idle timer is running, we're not supposed to do anything
	if !c.spawnIdleTimeout.Done() {
		return nil
	}

	// Wave ongoing: Check if creeps to spawn left
	if c.creepsSpawned < c.activeWave.TotalCreepsToSpawn {
		return c.spawnCreep()
	}
	// All creaps dead, start idle timeout & update status
	if !c.waveCleared && c.creepsAlive == 0 {
		log.Println("Wave cleared. Setting idle timeout")
		c.waveCleared = true
		c.spawnIdleTimeout.Set(idleTimeAfterWaveFinished)
		return nil
	}
	// Next wave ready: Spawn next wave, if cleared & timeout done
	if c.waveCleared {
		log.Println("Spawning next wave")
		return c.NextWave()
	}

	return nil
}

// Special entity remove callback required to keep track of alive creeps
func (c *BaseCreepManager) RemoveEntity(entity BaseEntity) error {
	// Remove npc from game world
	if err := c.entityManager.RemoveEntity(entity); err != nil {
		return err
	}
	c.creepsAlive--
	return nil
}

func (c *BaseCreepManager) Progress() hud.ProgressInfo {
	label := fmt.Sprintf("Wave %d", c.activeWave.Round)
	// While idle timer is active, show remaining timeout
	if !c.spawnIdleTimeout.Done() {
		label := "DAY"
		current := int((idleTimeAfterWaveFinished - c.spawnIdleTimeout.Elapsed()) / idleTimeAfterWaveFinished * 100)
		return hud.ProgressInfo{Min: 0, Max: 100, Current: current, Label: label}
	}
	// While wave is spawning show wave progress
	return hud.ProgressInfo{Min: 0, Max: c.activeWave.TotalCreepsToSpawn, Current: c.creepsSpawned, Label: label}
}

func (c *BaseCreepManager) Round() int { return c.activeWave.Round }
func (c *BaseCreepManager) SetProvider(p CreepProvider) error {
	c.creepProvider = p
	return c.NextWave()
}

func (c *BaseCreepManager) spawnCreep() error {
	// Timeout until creep spawn timer over
	if !c.creepSpawnTimeout.Done() {
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
	c.creepSpawnTimeout.Set(1 / c.activeWave.WaveTicksPerSecond)
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
	log.Printf("Starting wave %v...\n", c.activeWave)
	c.creepSpawnTimeout.Set(1 / wave.WaveTicksPerSecond)
	return nil
}
