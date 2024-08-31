package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type AnimationController interface {
	Draw(RenderingTarget, *cp.Shape, Orientation) error
	// Play the given animation once
	Play(animation string) error
	Loop(animation string) error
}

type AnimationAsset interface {
	DrawAnimationTile(t RenderingTarget, shape *cp.Shape, animation *GameAssetAnimation, animationTile int, o Orientation) error
	Animation(animation string) (*GameAssetAnimation, error)
	AnimationTimeProvider
}

type BaseAnimationManager struct {
	asset AnimationAsset

	playingAnimation      *GameAssetAnimation
	playingAnimationTimer Timer

	loopAnimation         *GameAssetAnimation
	loopingAnimationTimer Timer
}

func NewAnimationManager(asset AnimationAsset) (*BaseAnimationManager, error) {
	am := &BaseAnimationManager{}
	am.asset = asset

	var err error
	if am.playingAnimationTimer, err = NewAnimationTimer(asset); err != nil {
		return nil, err
	}
	if am.loopingAnimationTimer, err = NewAnimationTimer(asset); err != nil {
		return nil, err
	}
	if am.loopAnimation, err = asset.Animation("idle"); err != nil {
		return nil, fmt.Errorf("Cannot init animation manager without idle animation: %e", err.Error())
	}
	am.loopingAnimationTimer.Start()

	return am, nil
}

func (a *BaseAnimationManager) Play(animationKey string) error {
	var err error
	a.playingAnimation, err = a.asset.Animation(animationKey)
	if err != nil {
		return err
	}
	a.playingAnimationTimer.Stop()
	a.playingAnimationTimer.Start()
	return nil
}

func (a *BaseAnimationManager) Loop(animationKey string) error {
	var err error
	a.loopAnimation, err = a.asset.Animation(animationKey)
	if err != nil {
		return err
	}
	return nil
}

func (a *BaseAnimationManager) Draw(t RenderingTarget, shape *cp.Shape, o Orientation) error {
	// Check if there is an animation currently playing, if not, play loop
	if !a.playingAnimationTimer.Active() || a.playingAnimation == nil {
		currentAnimationTile := 0
		if a.loopAnimation == nil {
			return fmt.Errorf("Cannot draw: Neither play, nor loop animation defined")
		}
		if a.loopAnimation.FrameCount > 1 && a.loopAnimation.Speed > 0 {
			currentAnimationTile = int(a.loopingAnimationTimer.Elapsed()/a.loopAnimation.Speed) % a.loopAnimation.FrameCount
		}
		return a.asset.DrawAnimationTile(t, shape, a.loopAnimation, currentAnimationTile, o)
	}

	// Calculate how many frames the animation has to play
	totalAnimationFrames := a.playingAnimation.Speed * float64(a.playingAnimation.FrameCount)
	diff := a.playingAnimationTimer.Elapsed()

	// Not finished playing
	if diff < float64(totalAnimationFrames) {
		// Calculate current animation tile
		currentAnimationTile := int(diff / float64(a.playingAnimation.Speed))
		return a.asset.DrawAnimationTile(t, shape, a.playingAnimation, currentAnimationTile, o)
	}

	// Finished playing, back to loop
	a.playingAnimationTimer.Stop()
	a.loopingAnimationTimer.Start()
	return a.asset.DrawAnimationTile(t, shape, a.loopAnimation, 0, o)
}
