package td

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/hud"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type TDGame struct {
	world                     *engine.GameWorld
	screenWidth, screenHeight int
	creepManager              engine.CreepManager
	towerManager              *TowerManager
	goldManager               loot.ResourceManager
	hud                       *hud.GameHUD
	castle                    *CastleEntity
}

func (g *TDGame) Update() error {
	if g.world.IsOver() {
		// Wait for restart
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			err := g.initialize()
			if err != nil {
				fmt.Println("Could not restart game: ", err.Error())
			}
		}
		return nil
	}
	g.world.Update()
	if err := g.creepManager.Update(); err != nil {
		fmt.Println("Could not update creeps: ", err.Error())
	}
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

// Initialize all parts of the game world that need to be reset on restart
func (game *TDGame) initialize() error {
	fmt.Println("Initializing game")
	// Init game world
	width := int64(game.screenWidth)
	height := int64(game.screenHeight)
	w, err := engine.NewWorld(width, height)
	if err != nil {
		return err
	}
	am := w.AssetManager
	// Initialize map
	tileset, err := am.Tileset("plains")
	if err != nil {
		return err
	}
	w.WorldMap, err = engine.NewWorldMap(width, height)
	if err != nil {
		return err
	}
	if err = w.WorldMap.AddLayer(assets.LabyrinthMapCSV, tileset); err != nil {
		return err
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
	castleAsset, err := am.CharacterAsset("castle")
	if err != nil {
		return fmt.Errorf("Could not find castle asset")
	}
	game.castle, err = NewCastle(w, castleAsset, game.EndGame)
	if err != nil {
		return err
	}
	w.AddEntity(game.castle)

	// Setup gold management
	game.goldManager, err = loot.NewInMemoryResourceManager()
	if err != nil {
		return err
	}
	// Add starting gold
	game.goldManager.Add(50)

	// Setup tower management
	game.towerManager, err = NewTowerManager(w, am, game.goldManager, game.world.WorldMap)

	// Setup creep management
	npcAsset, err := w.AssetManager.CharacterAsset("npc-torch")
	if err != nil {
		return fmt.Errorf("Cannot initialize creep management: Could not find npc asset")
	}
	game.creepManager, err = engine.NewDefaultCreepManager(w, npcAsset)
	if err != nil {
		return err
	}

	// Setup HUD. Needs to be reset to initialize speed slider correctly
	game.hud, err = hud.NewHUD(game)
	if err != nil {
		return err
	}

	// Setup camera
	cam, err := engine.NewBaseCamera(game.screenWidth, game.screenHeight)
	if err != nil {
		return err
	}
	game.world.SetCamera(cam)

	return nil
}

// Constructor: Initialize parts of game that are constant even after restarting
func NewTDGame(screenWidth, screenHeight int) (*TDGame, error) {
	game := &TDGame{screenWidth: screenWidth, screenHeight: screenHeight}

	// Initialize
	if err := game.initialize(); err != nil {
		return nil, err
	}
	return game, nil
}

func (g *TDGame) GameOver() bool                   { return g.world.IsOver() }
func (g *TDGame) CreepProgress() hud.ProgressInfo  { return g.creepManager.Progress() }
func (g *TDGame) CastleProgress() hud.ProgressInfo { return g.castle.GetHealthBar() }
func (g *TDGame) SetSpeed(speed float64)           { g.world.GameSpeed = speed }
func (g *TDGame) Score() hud.ScoreValue {
	return hud.ScoreValue(float64(g.goldManager.Revenue()))
}

func (g *TDGame) EndGame() {
	g.world.EndGame()

	// Keeping score
	fmt.Printf("You've lost at wave %d \n", g.creepManager.Round())
	g.hud.SaveScore(g.Score())

	fmt.Println("Waiting for restart...")
}
