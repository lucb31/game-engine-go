package engine

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
)

type RenderingTarget interface {
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
	StrokeRect(x, y float64, width, height float32, strokeWidth float32, clr color.Color, antialias bool)
	StrokeCircle(cx, cy float64, r, strokeWidth float32, clr color.Color, antialias bool)
	Screen() *ebiten.Image
}

type Camera interface {
	RenderingTarget

	// Physics
	Position() cp.Vector
	Body() *cp.Body
	Shape() *cp.Shape

	// Drawing entities
	IsVisible(GameEntity) bool
	// Transforms world coordinates to camera coordinates
	WorldToScreenPos(cp.Vector) cp.Vector

	// General rendering
	DrawDebugInfo()
	// Call at start of every draw cycle
	SetScreen(*ebiten.Image)
	// Returns coordinates of viewport top-left & bottom-right vectors in world coordinates
	Viewport() (cp.Vector, cp.Vector)
	ViewportWidth() int
	ViewportHeight() int
}

type BaseCamera struct {
	viewportWidth, viewportHeight int
	shape                         *cp.Shape
	screen                        *ebiten.Image
}

func NewBaseCamera(width, height int) (*BaseCamera, error) {
	cam := &BaseCamera{viewportWidth: width, viewportHeight: height}
	// Init camera physics (required for cam movement)
	camBody := cp.NewKinematicBody()
	camBody.SetPosition(cp.Vector{X: float64(width) / 2, Y: float64(height) / 2})
	camBody.UserData = cam

	// Collision model
	cam.shape = cp.NewBox(camBody, float64(width), float64(height), 1)
	cam.shape.SetSensor(true)
	cam.shape.SetFilter(cp.SHAPE_FILTER_NONE)

	return cam, nil
}

func (c *BaseCamera) IsVisible(entity GameEntity) bool {
	return c.shape.BB().Intersects(entity.Shape().BB())
}

// Replacement for screen.DrawImage
// Translates image from absolute position to relative camera position
func (c *BaseCamera) DrawImage(im *ebiten.Image, op *ebiten.DrawImageOptions) {
	// NOTE: Creating COPY here. If we pass by reference here we'd change
	// the original opts which breaks cases where the same opts are re-used
	opts := *op
	opts.GeoM.Concat(c.worldMatrix())
	c.screen.DrawImage(im, &opts)
}

// Returns geometrix matrix that includes all camera translation, rotation, etc
func (c *BaseCamera) worldMatrix() ebiten.GeoM {
	res := ebiten.GeoM{}

	// Offset by camera viewport
	tL, _ := c.Viewport()
	res.Translate(-tL.X, -tL.Y)

	return res
}

// Draw stroked rectangle on provided absolute world position
func (c *BaseCamera) StrokeRect(absX, absY float64, width, height float32, strokeWidth float32, clr color.Color, antialias bool) {
	screen := c.screen
	tL, _ := c.Viewport()
	relX := float32(absX - tL.X)
	relY := float32(absY - tL.Y)
	vector.StrokeRect(screen, relX, relY, width, height, strokeWidth, clr, antialias)
}

func (c *BaseCamera) StrokeCircle(cx, cy float64, r, strokeWidth float32, clr color.Color, antialias bool) {
	screen := c.screen
	tL, _ := c.Viewport()
	relX := float32(cx - tL.X)
	relY := float32(cy - tL.Y)
	vector.StrokeCircle(screen, relX, relY, r, strokeWidth, clr, antialias)
}

func (c *BaseCamera) Viewport() (cp.Vector, cp.Vector) {
	topLeft := cp.Vector{
		X: c.Position().X - float64(c.viewportWidth/2),
		Y: c.Position().Y - float64(c.viewportHeight/2),
	}
	bottomRight := cp.Vector{
		X: c.Position().X + float64(c.viewportWidth/2),
		Y: c.Position().Y + float64(c.viewportHeight/2),
	}
	// NOTE: We're not blocking the camera from moving out of bounds right now
	// Easy fix could be to make the map a little bigger than the outer bounds
	return topLeft, bottomRight
}

func (c *BaseCamera) SetScreen(screen *ebiten.Image) { c.screen = screen }

func (c *BaseCamera) DrawDebugInfo() {
	screen := c.screen
	if screen == nil {
		log.Fatalln("Cannot draw debug info without screen")
		return
	}
	// Draw camera position
	tl, br := c.Viewport()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Viewport: (%.1f, %.1f) - (%.1f, %.1f)", tl.X, tl.Y, br.X, br.Y), 10, 10)

	// Draw cursor position
	relX, relY := ebiten.CursorPosition()
	worldPos := c.ScreenToWorldPos(cp.Vector{float64(relX), float64(relY)})
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cursor Screen(%d, %d), World(%.1f, %.1f)", relX, relY, worldPos.X, worldPos.Y), 10, 20)
}

func (c *BaseCamera) WorldToScreenPos(worldPos cp.Vector) cp.Vector {
	topLeft, _ := c.Viewport()
	return worldPos.Sub(topLeft)
}

func (c *BaseCamera) ScreenToWorldPos(screenPos cp.Vector) cp.Vector {
	topLeft, _ := c.Viewport()
	return screenPos.Add(topLeft)
}

// ////////
// Getters
// ////////
func (c *BaseCamera) Position() cp.Vector   { return c.shape.Body().Position() }
func (c *BaseCamera) Body() *cp.Body        { return c.shape.Body() }
func (c *BaseCamera) Shape() *cp.Shape      { return c.shape }
func (c *BaseCamera) Screen() *ebiten.Image { return c.screen }
func (c *BaseCamera) ViewportWidth() int    { return c.viewportWidth }
func (c *BaseCamera) ViewportHeight() int   { return c.viewportHeight }
