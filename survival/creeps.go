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

// TODO: Config file
var availableNpcs = []NpcType{
	{"npc-torch", engine.NpcOpts{BasePower: 30, BaseHealth: 60, BaseMovementSpeed: 75.0}},
	{"npc-orc", engine.NpcOpts{BasePower: 15, BaseHealth: 90, BaseMovementSpeed: 50.0}},
	{"npc-slime", engine.NpcOpts{BasePower: 50, BaseHealth: 30, BaseMovementSpeed: 25.0}},
}

func NewSurvCreepManager(em engine.GameEntityManager, target engine.GameEntity, am engine.AssetManager, goldManager engine.GoldManager) (*engine.BaseCreepManager, error) {
	cm, err := engine.NewBaseCreepManager(em, goldManager)
	if err != nil {
		return nil, err
	}
	provider, err := NewSurvCreepProvider(am, target)
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
}

func NewSurvCreepProvider(am engine.AssetManager, t engine.GameEntity) (*SurvCreepProvider, error) {
	return &SurvCreepProvider{assetManager: am, target: t}, nil
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
	opts.StartingPos, err = calcCreepSpawnPosition()
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
// TODO: Creeps cannot spawn too close to the player
func calcCreepSpawnPosition() (cp.Vector, error) {
	// BOUNDS: 530, 402 - 2410,1710
	boundsMinX, boundsMinY, boundsMaxX, boundsMaxY := 530.0, 402.0, 2410.0, 1710.0
	for tries := 0; tries < 10; tries++ {
		randX := rand.Float64()*(boundsMaxX-boundsMinX) + boundsMinX
		randY := rand.Float64()*(boundsMaxY-boundsMinY) + boundsMinY

		// Castle 1140, 402 - 1815, 402
		castleArea := image.Rect(1140, 402, 1815, 1160)
		spawnArea := image.Rect(int(randX), int(randY), int(randX)+1, int(randY)+1)
		if !spawnArea.In(castleArea) {
			return cp.Vector{X: randX, Y: randY}, nil
		}
		fmt.Println("Intruder in the castle. Retrying...", randX, randY)
	}
	return cp.Vector{}, fmt.Errorf("Could not find a spawn position. Max tries reached")
}

// Choose a random npc type to spawn next
func (p *SurvCreepProvider) nextNpcType() NpcType {
	idx := rand.IntN(len(availableNpcs))
	return availableNpcs[idx]
}
