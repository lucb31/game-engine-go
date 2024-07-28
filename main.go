package main

import (
	"fmt"

	"github.com/lucb31/animation-go/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1024 / 2
	screenHeight = 768 / 2
	tileSize     = 16
)

type Game struct {
	world *engine.GameWorld
}

func Init() (*Game, error) {
	world, err := engine.NewWorld(screenWidth/tileSize, screenHeight/tileSize)
	if err != nil {
		return nil, err
	}
	return &Game{world: world}, nil
}

func (g *Game) Update() error {
	g.world.Update()
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		fmt.Println("go up")
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Animate")

	// Init game
	g, err := Init()
	if err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		fmt.Println(err)
	}
}
