package survival

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/hud"
)

type SurvivalGame struct {
	world        *engine.GameWorld
	camera       engine.Camera
	goldManager  engine.GoldManager
	creepManager *engine.CreepManager

	hud                       *hud.GameHUD
	worldWidth, worldHeight   int64
	screenWidth, screenHeight int
}

func (g *SurvivalGame) Update() error {
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
	g.hud.Update()

	return nil
}

func (g *SurvivalGame) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
	g.hud.Draw(screen)
}

func (g *SurvivalGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

// Initialize map layers from CSV data
func (game *SurvivalGame) initMap() error {
	// Base layer
	baseTiles, err := game.world.AssetManager.Tileset("grounds")
	if err != nil {
		return err
	}
	worldMap, err := engine.NewWorldMap(game.worldWidth, game.worldHeight, assets.MapTDBaseGroundCSV, baseTiles)
	if err != nil {
		return err
	}
	game.world.WorldMap = worldMap

	// Inner walls layer
	wallTiles, err := game.world.AssetManager.Tileset("fences")
	if err != nil {
		return err
	}
	game.world.AddCollisionLayer(assets.MapTDBaseWallsCSV, wallTiles)

	return nil
}

// Initialize all parts of the game world that need to be reset on restart
func (game *SurvivalGame) initialize() error {
	game.worldHeight = int64(2112)
	game.worldWidth = int64(2944)

	fmt.Println("Initializing game")
	// Init game world
	w, err := engine.NewWorld(game.worldWidth, game.worldHeight)
	if err != nil {
		return err
	}
	game.world = w
	am := w.AssetManager

	// Initialize game map
	if err = game.initMap(); err != nil {
		return err
	}

	// Init player
	player, err := game.world.InitPlayer(am)
	if err != nil {
		return err
	}

	// Init main camera
	camera, err := engine.NewFollowingCamera(game.screenWidth, game.screenHeight, player)
	if err != nil {
		return err
	}
	game.world.SetCamera(camera)

	// Setup gold management
	game.goldManager, err = engine.NewInMemoryGoldManager()
	if err != nil {
		return err
	}
	// Add starting gold
	game.goldManager.Add(50)

	// Setup creep management
	npcAsset, err := w.AssetManager.CharacterAsset("npc-torch")
	if err != nil {
		return fmt.Errorf("Cannot initialize creep management: Could not find npc asset")
	}
	game.creepManager, err = NewSurvCreepManager(w, player, npcAsset, game.goldManager)
	if err != nil {
		return err
	}

	// Init hud
	game.hud, err = hud.NewHUD(game)
	if err != nil {
		return err
	}
	return nil
}

// Constructor: Initialize parts of game that are constant even after restarting
func NewSurvivalGame(screenWidth, screenHeight int) (*SurvivalGame, error) {
	game := &SurvivalGame{screenWidth: screenWidth, screenHeight: screenHeight}

	// Initialize
	if err := game.initialize(); err != nil {
		return nil, err
	}
	return game, nil
}

func (g *SurvivalGame) SetSpeed(speed float64)           { g.world.GameSpeed = speed }
func (g *SurvivalGame) GameOver() bool                   { return g.world.IsOver() }
func (g *SurvivalGame) Balance() int64                   { return g.goldManager.Balance() }
func (g *SurvivalGame) Score() hud.ScoreValue            { return hud.ScoreValue(g.goldManager.Revenue()) }
func (g *SurvivalGame) CastleProgress() hud.ProgressInfo { return hud.ProgressInfo{} }
func (g *SurvivalGame) CreepProgress() hud.ProgressInfo  { return g.creepManager.Progress() }

func (g *SurvivalGame) EndGame() {
	g.world.EndGame()
	fmt.Println("Waiting for restart...")
}
