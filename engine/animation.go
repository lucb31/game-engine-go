package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type AnimationController interface {
	Play(string, int) error
	Loop(string) error
	Draw(RenderingTarget, *cp.Shape) error
}

type BaseAnimationManager struct {
	asset *CharacterAsset

	playingAnimation string
	// Frame count when playing animation was started
	playingAnimationSinceFrame int64
	playingAnimationSpeed      int
	loopAnimation              string
}

func NewAnimationManager(asset *CharacterAsset) (*BaseAnimationManager, error) {
	am := &BaseAnimationManager{}
	am.asset = asset

	am.playingAnimation = "idle_east"
	return am, nil
}

// Play the given animation once with given speed
func (a *BaseAnimationManager) Play(animation string, speed int) error {
	_, ok := a.asset.Animations[animation]
	if !ok {
		return fmt.Errorf("Unknown animation: %s", animation)
	}
	fmt.Println("Gonna play", animation, speed)
	a.playingAnimation = animation
	a.playingAnimationSpeed = speed
	a.playingAnimationSinceFrame = *a.asset.currentFrame
	return nil
}

func (a *BaseAnimationManager) Loop(animation string) error {
	_, ok := a.asset.Animations[animation]
	if !ok {
		return fmt.Errorf("Unknown animation: %s", animation)
	}
	a.loopAnimation = animation
	return nil
}

func (a *BaseAnimationManager) Draw(t RenderingTarget, shape *cp.Shape) error {
	// Check if there is an animation currently playing, if not, play loop
	if a.playingAnimationSinceFrame == 0 {
		return a.asset.Draw(t, a.loopAnimation, shape)
	}

	// Calculate how many frames the animation has to play
	animation := a.asset.Animations[a.playingAnimation]
	animationSpeed := a.playingAnimationSpeed
	totalAnimationFrames := animationSpeed * animation.FrameCount
	diff := *a.asset.currentFrame - a.playingAnimationSinceFrame

	// Not finished playing
	if diff < int64(totalAnimationFrames) {
		// Calculate current animation tile
		currentAnimationTile := int(diff / int64(animationSpeed))
		return a.asset.DrawAnimationTile(t, animation, int(currentAnimationTile), shape)
	}

	// Finished playing, back to loop
	fmt.Println("finished playing", a.playingAnimation)
	a.playingAnimationSinceFrame = 0
	return a.asset.Draw(t, a.loopAnimation, shape)
}
