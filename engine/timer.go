package engine

import (
	"fmt"
	"math"
)

type Timer interface {
	Start()
	Stop()
	Active() bool
	Elapsed() float64
}

type TimeFunc func() float64

type BaseTimer struct {
	startedAt float64
	time      TimeFunc
}

func NewIngameTimer(p IngameTimeProvider) (*BaseTimer, error) {
	if p == nil {
		return nil, fmt.Errorf("Cannot init timer without time provider")
	}
	return &BaseTimer{time: p.IngameTime}, nil
}

func NewAnimationTimer(p AnimationTimeProvider) (*BaseTimer, error) {
	if p == nil {
		return nil, fmt.Errorf("Cannot init timer without time provider")
	}
	return &BaseTimer{time: p.AnimationTime}, nil
}

func (t *BaseTimer) Start() {
	// 0.0001 ensures that we dont run into inf elapsed time, but actual time dif
	t.startedAt = math.Max(t.time(), 0.0001)
}
func (t *BaseTimer) Stop()        { t.startedAt = 0 }
func (t *BaseTimer) Active() bool { return t.startedAt != 0 }
func (t *BaseTimer) Elapsed() float64 {
	// Return INF if not started
	if t.startedAt == 0 {
		return math.MaxFloat64
	}
	now := t.time()
	return now - t.startedAt
}
