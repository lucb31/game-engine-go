package engine

import (
	"log"
	"time"
)

type ExecutionDebugger struct {
	durations []time.Duration
	name      string
}

func NewExecutionDebugger(name string) *ExecutionDebugger {
	return &ExecutionDebugger{name: name, durations: make([]time.Duration, 0)}
}

func (e *ExecutionDebugger) AvgExeTime() func() {
	start := time.Now()
	return func() {
		e.durations = append(e.durations, time.Since(start))
		// Output every 120th iteration (should be once ever 2 second)
		if len(e.durations) >= 120 {
			sum := int64(0)
			for _, duration := range e.durations {
				sum += int64(duration)
			}
			avg := time.Duration(sum / int64(len(e.durations)))
			log.Printf("%s avg execution time: %v\n", e.name, avg)
			e.durations = make([]time.Duration, 0)
		}
	}
}

// Debug method to measure execution time of functions.
// Very useful for rendering optimizations
func ExeTime(name string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s execution time: %v\n", name, time.Since(start))
	}
}
