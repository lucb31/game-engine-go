package main

import (
	"fmt"
	"image"
	_ "image/png"

	"github.com/lucb31/animation-go/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth         = 1024 / 2
	screenHeight        = 768 / 2
	frameWidth          = 48
	frameHeight         = 48
	animationFrameCount = 6
	tileSize            = 16
)

type Game struct {
	frameCount   int64
	world        engine.GameWorld
	assetManager *engine.AssetManager
}

func Init() (*Game, error) {
	am, err := engine.NewAssetManager()
	if err != nil {
		return nil, err
	}
	world, err := engine.NewWorld(12, 8)
	if err != nil {
		return nil, err
	}
	return &Game{world: world, assetManager: am}, nil
}

func (g *Game) Update() error {
	g.frameCount++
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		// if keyPressed(g.keys, "ArrowUp") {
		fmt.Println("go up")
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBiomes(screen)
	g.drawPlayer(screen)
}

func (g *Game) drawBiomes(screen *ebiten.Image) {
	// Todo this doesnt change per frame
	// Todo Currently drawing WHOLE map. Should only draw visible map
	for row := range g.world.Height {
		for col := range g.world.Width {
			// Set tile position
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*tileSize), float64(row*tileSize))

			// Select correct tile from tileset
			tileIdx := int(g.world.Biome[row][col])
			tileX := tileIdx % g.assetManager.TilesPerRow
			tileY := int(tileIdx / g.assetManager.TilesPerRow)

			subIm := g.assetManager.PlainsTileset.SubImage(image.Rect(tileSize*tileX, tileSize*tileY, tileSize*(tileX+1), tileSize*(tileY+1))).(*ebiten.Image)
			screen.DrawImage(subIm, &op)

		}
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	// Position in the center of the screen
	op := ebiten.DrawImageOptions{}
	// Offset size of player frame
	op.GeoM.Translate(-frameWidth/2, -frameHeight/2)
	op.GeoM.Translate(tileSize*6, tileSize*4)

	animationFrame := int(g.frameCount/6) % animationFrameCount
	tilePosition := 4
	subIm := g.assetManager.PlayerTileset.SubImage(image.Rect(
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

func main() {
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
