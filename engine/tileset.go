package engine

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Tileset struct {
	images []*ebiten.Image
}

func NewTileset(tilesetImage *ebiten.Image, tilesPerRow int, tileSize int, scale float64) (*Tileset, error) {
	tilesSetSize := tilesPerRow * tileSize
	images := make([]*ebiten.Image, tilesSetSize)

	for tileIdx := range tilesSetSize {
		tileX := tileIdx % tilesPerRow
		tileY := int(tileIdx / tilesPerRow)
		// Selecting sub image based on tile information
		rawIm := tilesetImage.SubImage(image.Rect(
			tileX*tileSize,
			tileY*tileSize,
			(tileX+1)*tileSize,
			(tileY+1)*tileSize,
		)).(*ebiten.Image)
		images[tileIdx] = scaleImg(rawIm, scale)
	}
	return &Tileset{images: images}, nil
}

func scaleImg(im *ebiten.Image, scale float64) *ebiten.Image {
	if scale == 1.0 || im.Bounds().Dx() == 0 || im.Bounds().Dy() == 0 {
		return im
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	result := ebiten.NewImage(int(float64(im.Bounds().Dx())*scale), int(float64(im.Bounds().Dy())*scale))
	result.DrawImage(im, &op)
	return result
}

func (t *Tileset) GetTile(tileIdx int) (*ebiten.Image, error) {
	if tileIdx < 0 || tileIdx > len(t.images)-1 {
		return nil, fmt.Errorf("Tileset out of bounds! Unknown index %d", tileIdx)
	}
	return t.images[tileIdx], nil
}
