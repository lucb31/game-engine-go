package engine

import (
	"testing"

	"github.com/jakecoffman/cp"
)

type SnapTest struct {
	input  cp.Vector
	output cp.Vector
}

func TestSnapToGrid(t *testing.T) {
	testCases := []SnapTest{
		{cp.Vector{0, 0}, cp.Vector{16, 16}},
		{cp.Vector{5, 0}, cp.Vector{16, 16}},
		{cp.Vector{5, 9}, cp.Vector{16, 16}},
		{cp.Vector{95, 55}, cp.Vector{80, 48}},
		{cp.Vector{96, 55}, cp.Vector{112, 48}},
	}
	for _, testCase := range testCases {
		res := SnapToGrid(testCase.input, 32, 32)
		if res != testCase.output {
			t.Fatal("Input != Output", testCase, res)
		}
	}
}
