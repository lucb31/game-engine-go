package survival

import (
	"fmt"
	"image"
	"math/rand/v2"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

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

func (p *SurvCreepProvider) NextNpc(remover engine.EntityRemover, opts engine.NpcOpts) (engine.GameEntity, error) {
	npcAsset, err := p.calcCreepAsset()
	if err != nil {
		return nil, err
	}
	opts.StartingPos, err = calcCreepSpawnPosition()
	if err != nil {
		return nil, err
	}
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

func (p *SurvCreepProvider) calcCreepAsset() (*engine.CharacterAsset, error) {
	availableAssets := []string{"npc-torch", "npc-orc", "npc-slime"}
	idx := rand.IntN(len(availableAssets))
	return p.assetManager.CharacterAsset(availableAssets[idx])
}
