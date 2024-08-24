package engine

import (
	"fmt"
	"math"
)

type IngameTimer struct {
	timeProvider IngameTimeProvider

	startedAt float64
}

func NewIngameTimer(p IngameTimeProvider) (*IngameTimer, error) {
	if p == nil {
		return nil, fmt.Errorf("Cannot init timer without time provider")
	}
	return &IngameTimer{timeProvider: p}, nil
}

func (t *IngameTimer) Start()       { t.startedAt = t.timeProvider.GetIngameTime() }
func (t *IngameTimer) Stop()        { t.startedAt = 0 }
func (t *IngameTimer) Active() bool { return t.startedAt != 0 }
func (t *IngameTimer) Elapsed() float64 {
	// Return INF if not started
	if t.startedAt == 0 {
		return math.MaxFloat64
	}
	now := t.timeProvider.GetIngameTime()
	return now - t.startedAt
}
