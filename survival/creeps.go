package survival

import (
	"fmt"
	"image"
	"math"
	"math/rand/v2"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/hud"
)

type CreepSpawner interface {
	Spawn()
}

type CreepManager struct {
	entityManager engine.GameEntityManager
	goldManager   engine.GoldManager
	asset         *engine.CharacterAsset

	activeWave *Wave

	// Active wave
	creepsSpawned        int
	creepsAlive          int
	lastCreepSpawnedTime float64

	// Targetted by npcs
	target engine.GameEntity
}

type Wave struct {
	Round                   int
	CreepsToSpawn           int
	CreepOpts               engine.NpcOpts
	CreepSpawnRatePerSecond float64
}

const goldPerKill = int64(2)

func NewCreepManager(em engine.GameEntityManager, target engine.GameEntity, asset *engine.CharacterAsset, goldManager engine.GoldManager) (*CreepManager, error) {
	if asset == nil || em == nil {
		return nil, fmt.Errorf("Invalid arguments")
	}
	cm := &CreepManager{entityManager: em, asset: asset, goldManager: goldManager}
	cm.target = target
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
	var err error
	opts := c.activeWave.CreepOpts
	opts.StartingPos, err = calcCreepSpawnPosition()
	if err != nil {
		return err
	}
	npc, err := engine.NewNpcAggro(c, c.target, c.asset, opts)
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

// Creeps cannot spawn out of bounds
// Creeps cannot spawn within the castle area
// TODO: Creeps cannot spawn too close to the player
func calcCreepSpawnPosition() (cp.Vector, error) {
	// BOUNDS: 530, 402 - 2410,1710
	// Castle 1140, 402 - 1815, 402
	boundsMinX, boundsMinY, boundsMaxX, boundsMaxY := 530.0, 402.0, 2410.0, 1710.0
	for tries := 0; tries < 10; tries++ {
		randX := rand.Float64()*(boundsMaxX-boundsMinX) + boundsMinX
		randY := rand.Float64()*(boundsMaxY-boundsMinY) + boundsMinY

		castleArea := image.Rect(1140, 402, 1815, 1160)
		spawnArea := image.Rect(int(randX), int(randY), int(randX)+1, int(randY)+1)
		if !spawnArea.In(castleArea) {
			return cp.Vector{X: randX, Y: randY}, nil
		}
		fmt.Println("Intruder in the castle. Retrying...", randX, randY)
	}
	return cp.Vector{}, fmt.Errorf("Could not find a spawn position. Max tries reached")
}

// Wave scaling
func calculateWaveOpts(round int) Wave {
	wave := Wave{Round: round}
	wave.CreepsToSpawn = int(math.Exp(float64(round)/4) + 29)
	wave.CreepSpawnRatePerSecond = 1.0
	startingHealth := math.Pow(3.5*float64(round), 2) + 100
	wave.CreepOpts = engine.NpcOpts{
		StartingHealth: startingHealth,
	}
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
