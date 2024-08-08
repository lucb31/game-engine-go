package main

import (
	"fmt"

	"github.com/lucb31/game-engine-go/td"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1024 / 2
	screenHeight = 768 / 2
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Animate")

	// Init game
	g, err := td.NewTDGame(screenWidth, screenHeight)
	if err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		fmt.Println(err)
	}
}
