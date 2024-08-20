package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type AnimationController interface {
	Play(string) error
	Loop(string) error
	Draw(RenderingTarget, *cp.Shape) error
}

type BaseAnimationManager struct {
	asset *CharacterAsset

	playingAnimation string
	// Frame count when playing animation was started
	playingAnimationSinceFrame int64
	loopAnimation              string
}

func NewAnimationManager(asset *CharacterAsset) (*BaseAnimationManager, error) {
	am := &BaseAnimationManager{}
	am.asset = asset

	am.playingAnimation = "idle_east"
	return am, nil
}

// Play the given animation ONCE
func (a *BaseAnimationManager) Play(animation string) error {
	_, ok := a.asset.Animations[animation]
	if !ok {
		return fmt.Errorf("Unknown animation: %s", animation)
	}
	fmt.Println("Gonna play", animation)
	a.playingAnimation = animation
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
	// FIX: This needs to come from somewhere else
	animationSpeed := 2
	totalAnimationFrames := animationSpeed * animation.FrameCount
	diff := *a.asset.currentFrame - a.playingAnimationSinceFrame

	// Not finished playing
	if diff < int64(totalAnimationFrames) {
		// Calculate current animation tile
		currentAnimationTile := int(diff / int64(animationSpeed))
		fmt.Printf("diff %v, total %d, current %d\n", diff, totalAnimationFrames, currentAnimationTile)
		return a.asset.DrawAnimationTile(t, animation, int(currentAnimationTile), shape)
	}

	// Finished playing, back to loop
	fmt.Println("finished playing", a.playingAnimation)
	a.playingAnimationSinceFrame = 0
	return a.asset.Draw(t, a.loopAnimation, shape)
}
