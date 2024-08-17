package main

import (
	"flag"
	"fmt"

	"github.com/lucb31/game-engine-go/survival"
	"github.com/lucb31/game-engine-go/td"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1024
	screenHeight = 768
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Game engine")

	// CLI
	var gameSelected string
	flag.StringVar(&gameSelected, "g", "survival", "Option to select game. Currently available 'td' & 'survival'")
	flag.Parse()

	// Init game
	var g ebiten.Game
	var err error
	switch gameSelected {
	case "td":
		g, err = td.NewTDGame(screenWidth, screenHeight)
	case "survival":
		g, err = survival.NewSurvivalGame(screenWidth, screenHeight)
	default:
		panic("No game found")
	}
	if err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		fmt.Println(err)
	}
}
