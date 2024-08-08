package td

import (
	"fmt"

	"github.com/lucb31/game-engine-go/assets"
	"github.com/lucb31/game-engine-go/engine"
)

func NewTDWorld(width int64, height int64) (*engine.GameWorld, error) {
	w, err := engine.NewWorld(width, height)
	if err != nil {
		return nil, err
	}
	am := w.AssetManager
	// Initialize map
	w.WorldMap, err = engine.NewWorldMap(width, height, assets.TestMapCSV, am.Tilesets["plains"])
	if err != nil {
		return nil, err
	}

	// Initialize an npc
	npcAsset, ok := am.CharacterAssets["npc-torch"]
	if !ok {
		return nil, fmt.Errorf("Could not find npc asset")
	}
	npc, err := engine.NewNpc(w, &npcAsset)
	if err != nil {
		return w, err
	}
	w.AddEntity(npc)

	// Initialize a tower
	towerAsset, ok := am.CharacterAssets["tower-blue"]
	if !ok {
		return nil, fmt.Errorf("Could not find tower asset")
	}
	projectile, ok := am.ProjectileAssets["bone"]
	if !ok {
		return nil, fmt.Errorf("Could not find projectile asset")
	}
	tower, err := NewTower(w, &towerAsset, &projectile)
	if err != nil {
		return w, err
	}
	w.AddEntity(tower)
	return w, nil
}
