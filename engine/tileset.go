package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Tileset struct {
	Image       *ebiten.Image
	TilesPerRow int
	TileSize    int
}

// NOTE: It might be more efficient to generate the subimages once and then just retrieve them
func (t *Tileset) GetTile(tileIdx int) (*ebiten.Image, error) {
	tileX := tileIdx % t.TilesPerRow
	tileY := int(tileIdx / t.TilesPerRow)
	// Selecting sub image based on tile information
	return t.Image.SubImage(image.Rect(
		tileX*t.TileSize,
		tileY*t.TileSize,
		(tileX+1)*t.TileSize,
		(tileY+1)*t.TileSize,
	)).(*ebiten.Image), nil
}
