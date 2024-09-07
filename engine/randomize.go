package engine

import (
	"math/rand"
)

// Returns a list of indices with size "sampleSize" and the given probability distribution
func SampleWithRelativeProbabilities(probabilities []int, sampleSize int) []int {
	// Calculate total sum of probabilities (required for normalization)
	sum := 0
	for _, prob := range probabilities {
		sum += prob
	}

	// Calculate normalized pdf & cumulative probabilities
	pdf := make([]float32, len(probabilities))
	cdf := make([]float32, len(probabilities))
	pdf[0] = float32(probabilities[0]) / float32(sum)
	cdf[0] = pdf[0]
	for i := 1; i < len(probabilities); i++ {
		pdf[i] = float32(probabilities[i]) / float32(sum)
		cdf[i] = cdf[i-1] + pdf[i]
	}

	// Sample
	res := make([]int, sampleSize)
	for i := range sampleSize {
		randNumber := rand.Float32()
		bucket := 0
		for randNumber > cdf[bucket] {
			bucket++
		}
		res[i] = bucket
	}
	return res
}
