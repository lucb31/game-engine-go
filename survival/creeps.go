package survival

import (
	"fmt"
	"image"
	"math/rand/v2"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type SurvCreepProvider struct {
	asset  *engine.CharacterAsset
	target engine.GameEntity
}

func NewSurvCreepProvider(asset *engine.CharacterAsset, t engine.GameEntity) (*SurvCreepProvider, error) {
	return &SurvCreepProvider{asset: asset, target: t}, nil
}

func (p *SurvCreepProvider) NextNpc(remover engine.EntityRemover, opts engine.NpcOpts) (engine.GameEntity, error) {
	var err error
	opts.StartingPos, err = calcCreepSpawnPosition()
	if err != nil {
		return nil, err
	}
	npc, err := engine.NewNpcAggro(remover, p.target, p.asset, opts)
	if err != nil {
		return nil, err
	}
	return npc, nil
}

func NewSurvCreepManager(em engine.GameEntityManager, target engine.GameEntity, asset *engine.CharacterAsset, goldManager engine.GoldManager) (*engine.CreepManager, error) {
	cm, err := engine.NewCreepManager(em, asset, goldManager)
	if err != nil {
		return nil, err
	}
	provider, err := NewSurvCreepProvider(asset, target)
	if err != nil {
		return nil, err
	}
	cm.SetProvider(provider)
	return cm, nil
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
