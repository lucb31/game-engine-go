package engine

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Tileset struct {
	images []*ebiten.Image
}

func NewTileset(tilesetImage *ebiten.Image, tileSizeX, tileSizeY int, scale float64) (*Tileset, error) {
	tileCols := int(tilesetImage.Bounds().Dx() / tileSizeX)
	tileRows := int(tilesetImage.Bounds().Dy() / tileSizeY)
	tilesSetSize := tileCols * tileRows
	images := make([]*ebiten.Image, tilesSetSize)

	tileIdx := 0
	for tileY := range tileRows {
		for tileX := range tileCols {
			// Selecting sub image based on tile information
			rawIm := tilesetImage.SubImage(image.Rect(
				tileX*tileSizeX,
				tileY*tileSizeY,
				(tileX+1)*tileSizeX,
				(tileY+1)*tileSizeY,
			)).(*ebiten.Image)
			images[tileIdx] = ScaleImg(rawIm, scale)
			tileIdx++
		}
	}
	return &Tileset{images: images}, nil
}

func (t *Tileset) GetTile(tileIdx int) (*ebiten.Image, error) {
	if tileIdx < 0 || tileIdx > len(t.images)-1 {
		return nil, fmt.Errorf("Tileset out of bounds! Unknown index %d", tileIdx)
	}
	return t.images[tileIdx], nil
}
