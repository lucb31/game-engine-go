package engine

import "github.com/hajimehoshi/ebiten/v2"

func FlipHorizontal(source *ebiten.Image) *ebiten.Image {
	result := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1)
	op.GeoM.Translate(float64(source.Bounds().Dx()), 0)
	result.DrawImage(source, op)
	return result
}

func ScaleImg(im *ebiten.Image, scale float64) *ebiten.Image {
	if scale == 1.0 || im.Bounds().Dx() == 0 || im.Bounds().Dy() == 0 {
		return im
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	result := ebiten.NewImage(int(float64(im.Bounds().Dx())*scale), int(float64(im.Bounds().Dy())*scale))
	result.DrawImage(im, &op)
	return result
}
