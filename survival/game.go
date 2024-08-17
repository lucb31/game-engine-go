package survival

import (
	"fmt"
	"image"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/assets"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/hud"
)

type SurvivalGame struct {
	world  *engine.GameWorld
	camera engine.Camera

	hud                       *hud.GameHUD
	worldWidth, worldHeight   int
	screenWidth, screenHeight int
}

func (g *SurvivalGame) Update() error {
	if g.world.IsOver() {
		return nil
	}
	g.world.Update()
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

// Initialize all parts of the game world that need to be reset on restart
func (game *SurvivalGame) initialize() error {
	worldHeight := int64(2112)
	worldWidth := int64(2944)

	fmt.Println("Initializing game")
	// Init game world
	w, err := engine.NewWorld(worldWidth, worldHeight)
	if err != nil {
		return err
	}
	am := w.AssetManager
	// Initialize map
	tileset, err := am.Tileset("grounds")
	if err != nil {
		return err
	}
	w.WorldMap, err = engine.NewWorldMap(worldWidth, worldHeight, assets.MapTDCSV, tileset)
	if err != nil {
		return err
	}
	game.world = w

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

	// Add some npcs to test rendering
	npcAsset, err := am.CharacterAsset("npc-torch")
	if err != nil {
		return err
	}
	for range 10 {
		pos, err := calcCreepSpawnPosition()
		if err != nil {
			return err
		}
		npc, err := engine.NewNpcAggro(w, player, npcAsset, engine.NpcOpts{StartingPos: pos})
		if err != nil {
			return err
		}
		w.AddEntity(npc)
	}

	// Init hud
	game.hud, err = hud.NewHUD(game)
	if err != nil {
		return err
	}
	return nil
}

// Creeps cannot spawn out of bounds
// Creeps cannot spawn within the castle area
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
func (g *SurvivalGame) Balance() int64                   { return 0 }
func (g *SurvivalGame) Score() hud.ScoreValue            { return 0 }
func (g *SurvivalGame) CastleProgress() hud.ProgressInfo { return hud.ProgressInfo{} }
func (g *SurvivalGame) CreepProgress() hud.ProgressInfo  { return hud.ProgressInfo{} }

func (g *SurvivalGame) EndGame() {
	g.world.EndGame()
	fmt.Println("Waiting for restart...")
}
