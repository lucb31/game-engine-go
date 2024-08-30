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

type Timeout interface {
	Set(seconds float64)
	Done() bool
	Timer
}

type BaseTimeout struct {
	*BaseTimer
	timeout float64
}

func NewIngameTimeout(p IngameTimeProvider) (*BaseTimeout, error) {
	timer, err := NewIngameTimer(p)
	if err != nil {
		return nil, err
	}
	timeout := &BaseTimeout{BaseTimer: timer}
	return timeout, nil
}

func (t *BaseTimeout) Set(timeout float64) {
	t.timeout = timeout
	t.Stop()
	t.Start()
}

func (t *BaseTimeout) Done() bool {
	if t.timeout == 0 {
		return true
	}
	if !t.Active() {
		return false
	}
	if t.Elapsed() > t.timeout {
		return true
	}
	return false
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
	// Dont overwrite if already active
	if t.Active() {
		return
	}
	// 0.0001 ensures that we dont run into inf elapsed time, but actual time dif
	t.startedAt = math.Max(t.time(), 0.0001)
}
func (t *BaseTimer) Stop()        { t.startedAt = 0 }
func (t *BaseTimer) Active() bool { return t.startedAt != 0 }
func (t *BaseTimer) Elapsed() float64 {
	if !t.Active() {
		return 0
	}
	now := t.time()
	return now - t.startedAt
}
