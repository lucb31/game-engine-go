package engine_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/lucb31/game-engine-go/engine"
)

func TestProbabilitySample(t *testing.T) {
	sampleSize := 10000
	pdf := []int{1, 1, 5, 10}

	res := engine.SampleWithRelativeProbabilities(pdf, sampleSize)
	if len(res) != sampleSize {
		t.Fatalf("Expected %d items, but only received array of size %d", sampleSize, len(res))
	}
	// Count frequency
	frequencies := make([]int, len(pdf))
	for _, item := range res {
		frequencies[item]++
	}
	fmt.Println("frequency", frequencies)
	// Normalize frequencies
	normalizedFrequencies := make([]float64, len(pdf))
	for idx, freq := range frequencies {
		normalizedFrequencies[idx] = float64(freq) / float64(sampleSize)
	}
	fmt.Println("Normalized frequency", normalizedFrequencies)

	// Asset frequencies
	if math.Abs(float64(frequencies[0]-frequencies[1])/float64(frequencies[0])) > 0.2 {
		t.Fatal("Expected max 10% difference between equal probabilities")
	}
	if math.Abs(float64(5*frequencies[0]-frequencies[2])/float64(frequencies[2])) > 0.2 {
		t.Fatal("Expected max 10% difference between 5x probabilities")
	}
	if math.Abs(float64(10*frequencies[0]-frequencies[3])/float64(frequencies[3])) > 0.2 {
		t.Fatal("Expected max 10% difference between 10x probabilities")
	}

}
