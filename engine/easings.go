package engine

import "math"

// Easing out exponentially. Used to smoothen acceleration
func EaseOutExpo(x float64) float64 {
	if x >= 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*x)
}

func EaseInOutCubic(x float64) float64 {
	if x >= 1 {
		return 1
	}
	if x < 0.5 {
		return 4 * x * x * x
	}
	return 1 - math.Pow(-2*x+2, 3)/2
}
