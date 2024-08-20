package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

// Parses user input and translates into player movement
type PlayerController interface {
	CalcVelocity(max float64, t float64) cp.Vector
}

type KeyboardPlayerController struct {
	movingEastSince  float64
	movingWestSince  float64
	movingNorthSince float64
	movingSouthSince float64
}

const rampUpTimeInSeconds = 0.5

func (c *KeyboardPlayerController) CalcVelocity(maxVelocity, gameTime float64) cp.Vector {
	// Reading inputs
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		c.movingSouthSince = 0
		if c.movingNorthSince == 0 {
			c.movingNorthSince = gameTime
		}
	} else {
		c.movingNorthSince = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		c.movingNorthSince = 0
		if c.movingSouthSince == 0 {
			c.movingSouthSince = gameTime
		}
	} else {
		c.movingSouthSince = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		c.movingWestSince = 0
		if c.movingEastSince == 0 {
			c.movingEastSince = gameTime
		}
	} else {
		c.movingEastSince = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		c.movingEastSince = 0
		if c.movingWestSince == 0 {
			c.movingWestSince = gameTime
		}
	} else {
		c.movingWestSince = 0
	}

	// Apply smoothened velocity
	vel := cp.Vector{
		X: smoothenVel(c.movingEastSince, gameTime, maxVelocity) - smoothenVel(c.movingWestSince, gameTime, maxVelocity),
		Y: smoothenVel(c.movingSouthSince, gameTime, maxVelocity) - smoothenVel(c.movingNorthSince, gameTime, maxVelocity),
	}

	return vel
}

// Smoothen input value by comparing time difference between start and end time to ramp up time
// and applying easing function
func smoothenVel(startAt, gameTime, maxVel float64) float64 {
	if startAt == 0 {
		return 0
	}
	diff := gameTime - startAt
	progressInRampUp := diff / rampUpTimeInSeconds
	smoothenedProgress := easeOutExpo(progressInRampUp)
	smoothenedVelocity := smoothenedProgress * maxVel
	return smoothenedVelocity
}

// Easing out exponentially. Used to smoothen acceleration
func easeOutExpo(x float64) float64 {
	if x >= 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*x)
}

func NewKeyboardPlayerController() (*KeyboardPlayerController, error) {
	return &KeyboardPlayerController{}, nil
}
