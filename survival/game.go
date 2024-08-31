package survival

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/hud"
)

type SurvivalGame struct {
	world        *engine.GameWorld
	creepManager engine.CreepManager
	castle       *CastleEntity

	hud                       *hud.GameHUD
	screenWidth, screenHeight int
	audioContext              *audio.Context
}

func (g *SurvivalGame) Update() error {
	g.world.Update()
	g.hud.Update()
	if g.world.IsOver() {
		// Wait for restart
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			err := g.initialize()
			if err != nil {
				log.Println("Could not restart game: ", err.Error())
			}
		}
		return nil
	}
	if err := g.creepManager.Update(); err != nil {
		log.Println("Could not update creeps: ", err.Error())
	}

	return nil
}

func (g *SurvivalGame) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
	g.hud.Draw(screen)
}

func (g *SurvivalGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

// Initialize all parts of the game world that need to be reset on restart
func (game *SurvivalGame) initialize() error {
	log.Println("Initializing game")
	// Setup level generator
	generator, err := NewSurvLevelGenerator()
	if err != nil {
		return err
	}
	// FIX: Hard-coded level dimension
	generator.SetWorldDimensions(2912, 2912)
	generator.SetScreenDimension(game.screenWidth, game.screenHeight)

	// Generate random level
	gameWorld, err := engine.NewGeneratedWorld(generator)
	if err != nil {
		return fmt.Errorf("Error during level generation: %s", err.Error())
	}
	game.world = gameWorld
	am := gameWorld.AssetManager

	// Init player
	player, err := game.world.InitPlayer(am)
	if err != nil {
		return err
	}
	// Init wood harvesting
	axe, err := engine.NewWoodHarvestingTool(game.world, player)
	if err != nil {
		return err
	}
	player.SetAxe(axe)
	player.Shape().Body().SetPosition(cp.Vector{1456, 1656})

	// Init main camera
	camera, err := engine.NewFollowingCamera(game.screenWidth, game.screenHeight)
	if err != nil {
		return err
	}
	camera.SetTarget(player)
	game.world.SetCamera(camera)

	// Castle
	if err := game.initCastle(camera); err != nil {
		return err
	}

	// Setup creep management (AFTER castle, so we can use it as target for npcs)
	game.creepManager, err = engine.NewBaseCreepManager(gameWorld)
	if err != nil {
		return err
	}
	provider, err := NewSurvCreepProvider(am, game.castle, camera)
	if err != nil {
		return err
	}
	if err = provider.ParseNoSpawnArea(game.world.Width, game.world.Height, assets.MapDarkLogicSpawnAreaCSV); err != nil {
		return err
	}
	if err = provider.ParseCreepWaypoints(assets.MapDarkLogicWaypointsCSV, gameWorld.Space()); err != nil {
		return err
	}
	if err = game.creepManager.SetProvider(provider); err != nil {
		return err
	}

	// Init hud
	game.hud, err = game.initHud()
	if err != nil {
		return err
	}

	return nil
}

func (game *SurvivalGame) initCastle(camera *engine.FollowingCamera) error {
	var err error
	// Init castle
	game.castle, err = NewCastle(game.world, game.world.EndGame)
	if err != nil {
		return err
	}
	game.castle.Shape().Body().SetPosition(cp.Vector{1456, 1456})
	// Init asset
	castleAsset, err := game.world.AssetManager.CharacterAsset("castle")
	if err != nil {
		return err
	}
	game.castle.SetAsset(castleAsset)

	// Init gun
	gunOpts := engine.BasicGunOpts{FireRange: 512.0, FireRatePerSecond: 2.0}
	projAsset, err := game.world.AssetManager.ProjectileAsset("arrow")
	if err != nil {
		return err
	}
	gun, err := engine.NewAutoAimGun(game.world, game.castle, projAsset, gunOpts)
	if err != nil {
		return err
	}
	// Start with 3 projectiles
	gun.SetProjectileCount(3)
	game.castle.SetGun(gun)

	// NOTE: Required to allow camera to switch to castle when entering
	game.castle.SetCamera(camera)

	// Add to entity management
	game.world.AddEntity(game.castle)
	return nil
}

func (g *SurvivalGame) initHud() (*hud.GameHUD, error) {
	// Init base
	base, err := hud.NewHUD(g)
	if err != nil {
		return nil, err
	}

	// Init shop
	inventory := g.world.Player().Inventory()
	shop, err := NewShopMenu(inventory, g.world.Player(), g.castle)
	if err != nil {
		return nil, err
	}
	// Allow castle to control shop enabled state
	shop.SetShopEnabler(g.castle)
	shop.SetGunProvider(g.castle)
	base.AddSubMenu(shop)

	// Init inventory
	inventoryHud, err := hud.NewInventoryHud(inventory)
	if err != nil {
		return nil, err
	}
	base.AddSubMenu(inventoryHud)

	return base, nil
}

// Constructor: Initialize parts of game that are constant even after restarting
func NewSurvivalGame(screenWidth, screenHeight int) (*SurvivalGame, error) {
	game := &SurvivalGame{screenWidth: screenWidth, screenHeight: screenHeight}

	// Setup audio context
	game.audioContext = audio.NewContext(48000)

	// Initialize
	if err := game.initialize(); err != nil {
		return nil, err
	}

	return game, nil
}

func (g *SurvivalGame) SetSpeed(speed float64) { g.world.GameSpeed = speed }
func (g *SurvivalGame) GameOver() bool         { return g.world.IsOver() }
func (g *SurvivalGame) Score() hud.ScoreValue {
	return hud.ScoreValue(g.world.Player().Inventory().GoldManager().Revenue())
}
func (g *SurvivalGame) CastleProgress() hud.ProgressInfo { return g.castle.HealthBar() }
func (g *SurvivalGame) CreepProgress() hud.ProgressInfo  { return g.creepManager.Progress() }

func (g *SurvivalGame) EndGame() {
	g.world.EndGame()
	log.Println("Waiting for restart...")
}
