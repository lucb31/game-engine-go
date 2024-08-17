package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
)

type RenderingTarget interface {
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
	StrokeRect(x, y float64, width, height float32, strokeWidth float32, clr color.Color, antialias bool)
	StrokeCircle(cx, cy float64, r, strokeWidth float32, clr color.Color, antialias bool)
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
	AbsToRel(cp.Vector) cp.Vector

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
	cam.shape.SetFilter(cp.NewShapeFilter(0, cp.SHAPE_FILTER_NONE.Categories, OuterWallsCategory))

	return cam, nil
}

// Simple algorithm that checks if top left of entity bounding box is within cam bounds
func (c *BaseCamera) IsVisible(entity GameEntity) bool {
	// Calculate Bounding box top left corner position
	shape := entity.Shape()
	width := shape.BB().R - shape.BB().L
	height := shape.BB().T - shape.BB().B
	absBodyTopLeft := cp.Vector{
		X: shape.Body().Position().X - width/2,
		Y: shape.Body().Position().Y - height/2,
	}
	relBodyTopLeft := c.AbsToRel(absBodyTopLeft)

	// Check bounds
	if relBodyTopLeft.X < 0 || relBodyTopLeft.X > float64(c.ViewportWidth()) {
		return false
	}
	if relBodyTopLeft.Y < 0 || relBodyTopLeft.Y > float64(c.viewportHeight) {
		return false
	}
	return true
}

// Replacement for screen.DrawImage
// Translates image from absolute position to relative camera position
func (c *BaseCamera) DrawImage(im *ebiten.Image, op *ebiten.DrawImageOptions) {
	// NOTE: Creating COPY here. If we pass by reference here we'd change
	// the original opts which breaks cases where the same opts are re-used
	opts := *op
	// Offset by camera viewport
	tL, _ := c.Viewport()
	opts.GeoM.Translate(-tL.X, -tL.Y)
	c.screen.DrawImage(im, &opts)
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
		fmt.Println("No screen!")
		return
	}
	tl, br := c.Viewport()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Viewport: (%.1f, %.1f) - (%.1f, %.1f)", tl.X, tl.Y, br.X, br.Y), 10, 10)
}

func (c *BaseCamera) AbsToRel(absolutePos cp.Vector) cp.Vector {
	topLeft, _ := c.Viewport()
	return absolutePos.Sub(topLeft)
}

// ////////
// Getters
// ////////
func (c *BaseCamera) Position() cp.Vector { return c.shape.Body().Position() }
func (c *BaseCamera) Body() *cp.Body      { return c.shape.Body() }
func (c *BaseCamera) Shape() *cp.Shape    { return c.shape }
func (c *BaseCamera) ViewportWidth() int  { return c.viewportWidth }
func (c *BaseCamera) ViewportHeight() int { return c.viewportHeight }
