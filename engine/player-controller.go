package engine

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
)

// Parses user input and translates into player movement
type PlayerController interface {
	CalcVelocity(max float64) cp.Vector
	Orientation() Orientation
	// True WHILE interaction is ongoing
	Interacting() bool
	SetInteracting(bool)
	Update()
}

type KeyboardPlayerController struct {
	// Dependencies
	animationController AnimationController

	// General movement
	movingEastTimer  Timer
	movingWestTimer  Timer
	movingNorthTimer Timer
	movingSouthTimer Timer

	// Dash
	dashActiveTimer     Timer
	dashCooldownTimeout Timeout
	dashDirection       cp.Vector

	orientation Orientation

	interacting bool
}

func NewKeyboardPlayerController(ac AnimationController, igt IngameTimeProvider) (*KeyboardPlayerController, error) {
	c := &KeyboardPlayerController{}
	c.animationController = ac
	var err error
	if c.movingEastTimer, err = NewIngameTimer(igt); err != nil {
		return nil, err
	}
	if c.movingWestTimer, err = NewIngameTimer(igt); err != nil {
		return nil, err
	}
	if c.movingSouthTimer, err = NewIngameTimer(igt); err != nil {
		return nil, err
	}
	if c.movingNorthTimer, err = NewIngameTimer(igt); err != nil {
		return nil, err
	}
	if c.dashActiveTimer, err = NewIngameTimer(igt); err != nil {
		return nil, err
	}
	if c.dashCooldownTimeout, err = NewIngameTimeout(igt); err != nil {
		return nil, err
	}
	c.dashCooldownTimeout.Set(dashCooldownInSeconds)
	return c, nil
}

const (
	// CANNOT BE 0
	rampUpTimeInSeconds = 0.5

	// DASH
	dashDurationInSeconds = 0.2
	dashDistance          = 200.0
	dashCooldownInSeconds = 2.0
)

func (c *KeyboardPlayerController) Update() {
	// Reading movement inputs
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		c.movingSouthTimer.Stop()
		c.movingNorthTimer.Start()
	} else {
		c.movingNorthTimer.Stop()
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		c.movingNorthTimer.Stop()
		c.movingSouthTimer.Start()
	} else {
		c.movingSouthTimer.Stop()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		c.movingWestTimer.Stop()
		c.movingEastTimer.Start()
	} else {
		c.movingEastTimer.Stop()
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		c.movingEastTimer.Stop()
		c.movingWestTimer.Start()
	} else {
		c.movingWestTimer.Stop()
	}

	// Reading interaction inputs
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		c.interacting = true
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyE) {
		c.interacting = false
	}
}

func (c *KeyboardPlayerController) CalcVelocity(maxVelocity float64) cp.Vector {
	// Apply smoothened velocity
	vel := cp.Vector{
		X: smoothenVel(c.movingEastTimer.Elapsed(), maxVelocity) - smoothenVel(c.movingWestTimer.Elapsed(), maxVelocity),
		Y: smoothenVel(c.movingSouthTimer.Elapsed(), maxVelocity) - smoothenVel(c.movingNorthTimer.Elapsed(), maxVelocity),
	}

	// Add up velocity from walking & dashing
	totalVel := vel.Add(c.calcVelFromDash(vel))

	// Update orientation
	if totalVel.Length() > 0.0 {
		c.orientation = updateOrientation(c.orientation, totalVel)
	}
	animation := "idle"
	if vel.Length() > 5.0 {
		animation = "walk"
	}
	c.animationController.Loop(animation)

	return totalVel
}

func (c *KeyboardPlayerController) Interacting() bool {
	return c.interacting
}
func (c *KeyboardPlayerController) SetInteracting(val bool) { c.interacting = false }

func (c *KeyboardPlayerController) Orientation() Orientation { return c.orientation }

func (c *KeyboardPlayerController) calcVelFromDash(vel cp.Vector) cp.Vector {
	// Register new dashes
	if c.dashCooldownTimeout.Done() && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		fmt.Println("New dash queued")
		// While moving, dash in direction of movement
		// While standing still, dash in direction of last horizontal movement
		if vel.Length() > 0 {
			c.dashDirection = vel.Normalize()
		} else if c.orientation&West == 0 {
			c.dashDirection = cp.Vector{X: -1, Y: 0}
		} else {
			c.dashDirection = cp.Vector{X: 1, Y: 0}
		}

		// Queue animation
		c.animationController.Play("dash")
		c.dashActiveTimer.Start()
		c.dashCooldownTimeout.Set(dashCooldownInSeconds)
	}
	// Nothing to do if no dash ongoing
	if !c.dashActiveTimer.Active() {
		return cp.Vector{}
	}
	// Check if dashing finished
	if c.dashActiveTimer.Elapsed() > dashDurationInSeconds {
		c.dashActiveTimer.Stop()
		return cp.Vector{}
	}
	progressInRampUp := c.dashActiveTimer.Elapsed() / dashDurationInSeconds
	smoothenedProgress := easeInOutCubic(progressInRampUp)
	dashVelocity := dashDistance / dashDurationInSeconds
	smoothenedVelocity := smoothenedProgress * dashVelocity

	// Apply dashing velocity
	return c.dashDirection.Mult(smoothenedVelocity)
}

// Smoothen input value by comparing time difference between start and end time to ramp up time
// and applying easing function
func smoothenVel(elapsed, maxVel float64) float64 {
	if elapsed == math.MaxFloat64 {
		return 0
	}
	progressInRampUp := elapsed / rampUpTimeInSeconds
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
