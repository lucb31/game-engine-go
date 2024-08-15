package td

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/assets"
	"github.com/lucb31/game-engine-go/engine"
)

type TDGame struct {
	world                     *engine.GameWorld
	screenWidth, screenHeight int
	creepManager              *CreepManager
	towerManager              *TowerManager
	goldManager               engine.GoldManager
	hud                       *GameHUD
	castle                    *CastleEntity
}

func (g *TDGame) Update() error {
	if g.world.IsOver() {
		return nil
	}
	g.world.Update()
	g.creepManager.Update()
	g.towerManager.Update()
	g.hud.Update()

	return nil
}

func (g *TDGame) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
	g.towerManager.Draw(screen)
	g.hud.Draw(screen)
}

func (g *TDGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

func NewTDGame(screenWidth, screenHeight int) (*TDGame, error) {
	game := &TDGame{screenWidth: screenWidth, screenHeight: screenHeight}
	// Init game world
	width := int64(screenWidth)
	height := int64(screenHeight)
	w, err := engine.NewWorld(width, height)
	if err != nil {
		return nil, err
	}
	am := w.AssetManager
	// Initialize map
	w.WorldMap, err = engine.NewWorldMap(width, height, assets.LabyrinthMapCSV, am.Tilesets["plains"])
	if err != nil {
		return nil, err
	}
	game.world = w

	// Add collision handler for castle
	// TODO: Should be registered within castle
	handler := w.Space().NewCollisionHandler(cp.CollisionType(engine.NpcCollision), CastleCollision)
	handler.BeginFunc = func(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
		a, b := arb.Bodies()
		npc, ok := a.UserData.(*engine.NpcEntity)
		if !ok {
			return false
		}
		castle, ok := b.UserData.(*CastleEntity)
		if !ok {
			return false
		}
		castle.OnNpcHit(npc)
		return false
	}

	// Initialize castle
	castleAsset, ok := am.CharacterAssets["castle"]
	if !ok {
		return nil, fmt.Errorf("Could not find castle asset")
	}
	game.castle, err = NewCastle(w, &castleAsset)
	if err != nil {
		return nil, err
	}
	w.AddEntity(game.castle)

	// Setup gold management
	game.goldManager, err = engine.NewInMemoryGoldManager()
	if err != nil {
		return nil, err
	}
	// Add starting gold
	game.goldManager.Add(10)

	// Setup tower management
	towerAsset, ok := am.CharacterAssets["tower-blue"]
	if !ok {
		return nil, fmt.Errorf("Could not find tower asset")
	}
	projectile, ok := am.ProjectileAssets["bone"]
	if !ok {
		return nil, fmt.Errorf("Could not find projectile asset")
	}
	game.towerManager, err = NewTowerManager(w, &towerAsset, &projectile, game.goldManager)

	// Setup creep management
	npcAsset, ok := w.AssetManager.CharacterAssets["npc-torch"]
	if !ok {
		return nil, fmt.Errorf("Cannot initialize creep management: Could not find npc asset")
	}
	game.creepManager, err = NewCreepManager(w, &npcAsset, game.goldManager)
	if err != nil {
		return nil, err
	}

	// Setup HUD
	game.hud, err = NewHUD(game)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (g *TDGame) GetCreepProgress() ProgressInfo {
	return g.creepManager.GetProgress()
}

func (g *TDGame) GetCastleHealth() ProgressInfo {
	return g.castle.GetHealthBar()
}

func (g *TDGame) GetSpeed() float64 {
	return g.world.GameSpeed
}

func (g *TDGame) SetSpeed(speed float64) {
	g.world.GameSpeed = speed
}
