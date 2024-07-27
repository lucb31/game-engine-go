package engine

import "fmt"

type Biome int

const (
	Gras Biome = iota
	Rock
)

type GameWorld struct {
	objects []*GameObj
	player  *Player
	Biome   [][]Biome
	Width   int64
	Height  int64
}

func NewWorld(width int64, height int64) (GameWorld, error) {
	biome := [][]Biome{
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{13, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 15},
	}
	if int64(len(biome)) != height || int64(len(biome[0])) != width {
		return GameWorld{}, fmt.Errorf("Biome data not matching world size")
	}
	w := GameWorld{Biome: biome, Width: width, Height: height}
	return w, nil
}
