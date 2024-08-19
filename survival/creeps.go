package survival

import (
	"fmt"
	"image"
	"math/rand/v2"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type NpcType struct {
	assetName string
	opts      engine.NpcOpts
}

// ////////
// CONFIG
// ////////
const (
	boundsMinX, boundsMinY, boundsMaxX, boundsMaxY = 530.0, 402.0, 2410.0, 1710.0
)

// Map waypoints used to unstuck creeps
var aiWaypoints = []cp.Vector{
	{1465, 1000},
	{1465, 1185},
	{1140, 1185},
	{1810, 1185},
	{1140, 480},
	{1810, 480},
}

// TODO: Config file
var availableNpcs = []NpcType{
	{"npc-torch", engine.NpcOpts{BasePower: 30, BaseHealth: 60, BaseMovementSpeed: 75.0, GoldValue: 2, Waypoints: aiWaypoints}},
	{"npc-orc", engine.NpcOpts{BasePower: 15, BaseHealth: 90, BaseMovementSpeed: 50.0, GoldValue: 3, Waypoints: aiWaypoints}},
	{"npc-slime", engine.NpcOpts{BasePower: 50, BaseHealth: 30, BaseMovementSpeed: 25.0, GoldValue: 1, Waypoints: aiWaypoints}},
}

func NewSurvCreepManager(em engine.GameEntityManager, target engine.GameEntity, cam engine.Camera, am engine.AssetManager, goldManager engine.GoldManager) (*engine.BaseCreepManager, error) {
	cm, err := engine.NewBaseCreepManager(em, goldManager)
	if err != nil {
		return nil, err
	}
	provider, err := NewSurvCreepProvider(am, target, cam)
	if err != nil {
		return nil, err
	}
	cm.SetProvider(provider)
	err = cm.NextWave()
	if err != nil {
		return nil, err
	}
	return cm, nil
}

type SurvCreepProvider struct {
	assetManager engine.AssetManager
	target       engine.GameEntity
	// Required to not spawn within current viewport
	camera engine.Camera
}

func NewSurvCreepProvider(am engine.AssetManager, t engine.GameEntity, cam engine.Camera) (*SurvCreepProvider, error) {
	return &SurvCreepProvider{assetManager: am, target: t, camera: cam}, nil
}

func (p *SurvCreepProvider) NextNpc(remover engine.EntityRemover, wave engine.Wave) (engine.GameEntity, error) {
	npcType := p.nextNpcType()
	// Load asset
	npcAsset, err := p.assetManager.CharacterAsset(npcType.assetName)
	if err != nil {
		return nil, err
	}
	// Load opts & calculate starting position
	opts := npcType.opts
	opts.StartingPos, err = calcCreepSpawnPosition(p.camera)
	if err != nil {
		return nil, err
	}
	// Apply scaling
	opts.BaseHealth = wave.HealthScalingFunc(opts.BaseHealth)

	// Init npc
	npc, err := engine.NewNpcAggro(remover, p.target, npcAsset, opts)
	if err != nil {
		return nil, err
	}
	return npc, nil
}

// Creeps cannot spawn out of bounds
// Creeps cannot spawn within the castle area
// Creeps cannot spawn within visble camera viewport
// TODO: Smarter algorithm that defines valid spawn areas first and then places a point within them
func calcCreepSpawnPosition(cam engine.Camera) (cp.Vector, error) {
	for tries := 0; tries < 10; tries++ {
		// Generate random coordinates within bounds
		randX := rand.Float64()*(boundsMaxX-boundsMinX) + boundsMinX
		randY := rand.Float64()*(boundsMaxY-boundsMinY) + boundsMinY
		spawnArea := image.Rect(int(randX), int(randY), int(randX)+1, int(randY)+1)

		// Check if within camera viewport
		topLeft, bottomRight := cam.Viewport()
		cameraArea := image.Rect(int(topLeft.X), int(topLeft.Y), int(bottomRight.X), int(bottomRight.Y))
		if spawnArea.In(cameraArea) {
			fmt.Println("Position within viewport. Retrying...", randX, randY, topLeft, bottomRight)
			continue
		}

		// Check if within castle area
		castleArea := image.Rect(1140, 402, 1815, 1160)
		if spawnArea.In(castleArea) {
			fmt.Println("Position within castle. Retrying...", randX, randY)
			continue
		}
		return cp.Vector{X: randX, Y: randY}, nil
	}
	return cp.Vector{}, fmt.Errorf("Could not find a spawn position. Max tries reached")
}

// Choose a random npc type to spawn next
func (p *SurvCreepProvider) nextNpcType() NpcType {
	idx := rand.IntN(len(availableNpcs))
	return availableNpcs[idx]
}
