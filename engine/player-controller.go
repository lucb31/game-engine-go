package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

// Parses user input and translates into player movement
type PlayerController interface {
	CalcVelocity(cur cp.Vector, max float64) cp.Vector
}

type KeyboardPlayerController struct{}

func (c *KeyboardPlayerController) CalcVelocity(velocity cp.Vector, playerVelocity float64) cp.Vector {
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		velocity.Y = max(-playerVelocity, velocity.Y-playerVelocity*0.1)
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		velocity.Y = min(playerVelocity, velocity.Y+playerVelocity*0.1)
	} else {
		velocity.Y = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		velocity.X = -playerVelocity
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		velocity.X = playerVelocity
	} else {
		velocity.X = 0
	}
	return velocity
}

func NewKeyboardPlayerController() (*KeyboardPlayerController, error) {
	return &KeyboardPlayerController{}, nil
}
