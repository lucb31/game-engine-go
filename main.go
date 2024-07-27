package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth         = 1024 / 2
	screenHeight        = 768 / 2
	frameWidth          = 48
	frameHeight         = 48
	animationFrameCount = 6
)

// Assets
var playerImage *ebiten.Image

type Game struct {
	frameCount int64
}

func (g *Game) Update() error {
	g.frameCount++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	// Position in the center of the screen
	op.GeoM.Translate(-frameWidth/2, -frameHeight/2)
	op.GeoM.Translate(screenWidth/2, screenHeight/2)

	animationFrame := int(g.frameCount/6) % animationFrameCount
	tilePosition := 4
	subIm := playerImage.SubImage(image.Rect(
		animationFrame*frameWidth,
		tilePosition*frameHeight,
		(animationFrame+1)*frameWidth,
		(tilePosition+1)*frameHeight,
	)).(*ebiten.Image)
	screen.DrawImage(subIm, &op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func readPngAsset(path string) (image.Image, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewReader(dat))
	if err != nil {
		return nil, err
	}
	return im, nil
}

func main() {
	// Load assets
	var err error
	im, err := readPngAsset("assets/player.png")
	if err != nil {
		fmt.Println("Could not read assets!", err.Error())
	}
	playerImage = ebiten.NewImageFromImage(im)
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Animate")

	if err := ebiten.RunGame(&Game{}); err != nil {
		fmt.Println(err)
	}
}
