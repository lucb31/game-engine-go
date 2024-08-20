package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
)

// Parses user input and translates into player movement
type PlayerController interface {
	CalcVelocity(max, t float64) cp.Vector
}

type KeyboardPlayerController struct {
	movingEastSince  float64
	movingWestSince  float64
	movingNorthSince float64
	movingSouthSince float64

	dashingSince float64
}

const (
	// CANNOT BE 0
	rampUpTimeInSeconds = 0.5

	// DASH
	dashDurationInSeconds = 0.3
	dashVelocity          = 600.0
	dashCooldownInSeconds = 2.0
)

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

	totalVel := vel.Add(c.calcVelFromDash(gameTime))

	return totalVel
}

func (c *KeyboardPlayerController) calcVelFromDash(gameTime float64) cp.Vector {
	diff := gameTime - c.dashingSince
	// Register new dashes
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		// Timeout
		if diff > dashCooldownInSeconds {
			c.dashingSince = gameTime
			diff = 0
		}
	}
	// Nothing to do if no dash ongoing
	if c.dashingSince == 0 {
		return cp.Vector{}
	}
	// Check if dashing finished
	if diff > dashDurationInSeconds {
		c.dashingSince = 0
		return cp.Vector{}
	}
	progressInRampUp := diff / dashDurationInSeconds
	smoothenedProgress := easeInOutCubic(progressInRampUp)
	smoothenedVelocity := smoothenedProgress * dashVelocity
	// FIX: Dash direction (mising S & N + not working when standing still)
	if c.movingWestSince > 0 {
		smoothenedVelocity *= -1
	}

	// Apply dashing velocity
	return cp.Vector{
		X: smoothenedVelocity,
		Y: 0,
	}
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

func easeInOutCubic(x float64) float64 {
	if x >= 1 {
		return 1
	}
	if x < 0.5 {
		return 4 * x * x * x
	}
	return 1 - math.Pow(-2*x+2, 3)/2
}

func NewKeyboardPlayerController() (*KeyboardPlayerController, error) {
	return &KeyboardPlayerController{}, nil
}
