package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/lucb31/animation-go/engine"

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

var plainsTileset *ebiten.Image
var playerImage *ebiten.Image

type Game struct {
	frameCount int64
	world      engine.GameWorld
}

func Init() (*Game, error) {
	world, err := engine.NewWorld(12, 8)
	if err != nil {
		return nil, err
	}
	return &Game{world: world}, nil
}

func (g *Game) Update() error {
	g.frameCount++
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		// if keyPressed(g.keys, "ArrowUp") {
		fmt.Println("go up")
	}

	return nil
}

func keyPressed(haystack []ebiten.Key, needle string) bool {
	for _, val := range haystack {
		if val.String() == needle {
			return true
		}
	}
	return false
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBiomes(screen)
	g.drawPlayer(screen)
}

const tileSize = 16

func (g *Game) drawBiomes(screen *ebiten.Image) {
	// Todo this doesnt change per frame
	tilesPerRow := int(plainsTileset.Bounds().Dx() / tileSize)
	// Todo Currently drawing WHOLE map. Should only draw visible map
	for row := range g.world.Height {
		for col := range g.world.Width {
			// Set tile position
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*tileSize), float64(row*tileSize))

			// Select correct tile from tileset
			tileIdx := int(g.world.Biome[row][col])
			tileX := tileIdx % tilesPerRow
			tileY := int(tileIdx / tilesPerRow)

			// 0 => 0, 0
			// 2 => 32, 0
			// 4 => 64, 0
			// 6 => 96, 0
			// 7 => 0, 16
			subIm := plainsTileset.SubImage(image.Rect(tileSize*tileX, tileSize*tileY, tileSize*(tileX+1), tileSize*(tileY+1))).(*ebiten.Image)
			screen.DrawImage(subIm, &op)

		}
	}
	// g.world.Biom
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	// Position in the center of the screen
	op := ebiten.DrawImageOptions{}
	// Offset size of player frame
	op.GeoM.Translate(-frameWidth/2, -frameHeight/2)
	op.GeoM.Translate(tileSize*6, tileSize*4)

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

func ReadPngAsset(path string) (image.Image, error) {
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
	im, err := ReadPngAsset("assets/player.png")
	if err != nil {
		fmt.Println("Could not read assets!", err.Error())
	}
	playerImage = ebiten.NewImageFromImage(im)
	im, err = ReadPngAsset("assets/plains.png")
	if err != nil {
		fmt.Println("Could not read plains asset!", err.Error())
	}
	plainsTileset = ebiten.NewImageFromImage(im)
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Animate")

	// Init game
	g, err := Init()
	if err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		fmt.Println(err)
	}
}
