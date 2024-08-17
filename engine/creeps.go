package engine

import (
	"fmt"
	"math"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/hud"
)

type CreepManager struct {
	entityManager GameEntityManager
	goldManager   GoldManager
	asset         *CharacterAsset

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

func NewCreepManager(em GameEntityManager, asset *CharacterAsset, goldManager GoldManager) (*CreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	cm := &CreepManager{entityManager: em, asset: asset, goldManager: goldManager}
	cm.nextWave()
	return cm, nil
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
	npc, err := NewNpc(c, c.asset, c.activeWave.CreepOpts)
	if err != nil {
		return err
	}
	c.entityManager.AddEntity(npc)
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
	wave.CreepOpts = NpcOpts{
		StartingHealth: startingHealth,
		Waypoints: []cp.Vector{
			{X: 48, Y: 720},
			{X: 976, Y: 720},
			{X: 976, Y: 48},
			{X: 208, Y: 48},
			{X: 208, Y: 560},
			{X: 816, Y: 560},
			{X: 816, Y: 208},
			{X: 368, Y: 208},
			{X: 368, Y: 384},
			{X: 640, Y: 384},
		}}
	return wave
}

func (c *CreepManager) nextWave() error {
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

func (c *CreepManager) RemoveEntity(entity GameEntity) error {
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
	return c.activeWave.CreepSpawnRatePerSecond
}

func (c *CreepManager) creepsToSpawn() int {
	return c.activeWave.CreepsToSpawn
}

func (c *CreepManager) GetProgress() hud.ProgressInfo {
	label := fmt.Sprintf("Wave %d", c.activeWave.Round)
	return hud.ProgressInfo{Min: 0, Max: c.creepsToSpawn(), Current: c.creepsSpawned, Label: label}
}

func (c *CreepManager) Round() int {
	return c.activeWave.Round
}
