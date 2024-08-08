package td

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucb31/game-engine-go/assets"
	"github.com/lucb31/game-engine-go/engine"
)

type TDGame struct {
	world                     *engine.GameWorld
	screenWidth, screenHeight int
	lastCreepSpawned          time.Time
}

func (g *TDGame) Update() error {
	g.world.Update()

	return nil
}

func (g *TDGame) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
}

func (g *TDGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

func NewTDGame(screenWidth, screenHeight int) (*TDGame, error) {
	width := int64(screenWidth)
	height := int64(screenHeight)
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
		return nil, err
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
		return nil, err
	}
	w.AddEntity(tower)
	return &TDGame{world: w, screenWidth: screenWidth, screenHeight: screenHeight}, nil
}
