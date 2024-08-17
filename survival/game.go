package survival

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/assets"
	"github.com/lucb31/game-engine-go/engine"
)

type SurvivalGame struct {
	world                     *engine.GameWorld
	camera                    engine.Camera
	worldWidth, worldHeight   int
	screenWidth, screenHeight int
}

func (g *SurvivalGame) Update() error {
	if g.world.IsOver() {
		return nil
	}
	g.world.Update()

	return nil
}

func (g *SurvivalGame) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
}

func (g *SurvivalGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

// Initialize all parts of the game world that need to be reset on restart
func (game *SurvivalGame) initialize() error {
	worldHeight := int64(1600)
	worldWidth := int64(2024)

	fmt.Println("Initializing game")
	// Init game world
	w, err := engine.NewWorld(worldWidth, worldHeight)
	if err != nil {
		return err
	}
	am := w.AssetManager
	// Initialize map
	tileset, err := am.Tileset("plains")
	if err != nil {
		return err
	}
	w.WorldMap, err = engine.NewWorldMap(worldWidth, worldHeight, assets.LargeMapCSV, tileset)
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
	for i := range 10 {
		npc, err := engine.NewNpc(w, npcAsset, engine.NpcOpts{StartingPos: cp.Vector{float64(i * 200), float64(i * 200)}})
		if err != nil {
			return err
		}
		w.AddEntity(npc)
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

func (g *SurvivalGame) GetSpeed() float64 { return g.world.GameSpeed }

func (g *SurvivalGame) SetSpeed(speed float64) { g.world.GameSpeed = speed }

func (g *SurvivalGame) EndGame() {
	g.world.EndGame()
	fmt.Println("Waiting for restart...")
}
