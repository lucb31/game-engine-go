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
	StrokeRect(topLeftWorldPos, botRightWorldPos cp.Vector, strokeWidth float32, clr color.Color, antialias bool)
	FillRect(topLeftWorldPos, botRightWorldPos cp.Vector, clr color.Color, antialias bool)
	StrokeCircle(center cp.Vector, r, strokeWidth float32, clr color.Color, antialias bool)
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
	VectorVisible(cp.Vector) bool
	// Transforms world coordinates to camera coordinates
	WorldToScreenPos(cp.Vector) cp.Vector

	// General rendering
	DrawDebugInfo()
	// Call at start of every draw cycle
	SetScreen(*ebiten.Image)
	// Returns coordinates of viewport top-left & bottom-right vectors in world coordinates
	Viewport() (cp.Vector, cp.Vector)
	ScreenWidth() int
	ScreenHeight() int
	Zoom() float64
}

const (
	camZoomFactorMax     = 3.0
	camZoomFactorMin     = 0.5
	camZoomFactorDefault = 1.0
)

type BaseCamera struct {
	screenWidth, screenHeight int
	shape                     *cp.Shape
	screen                    *ebiten.Image

	ZoomFactor float64
}

func NewBaseCamera(width, height int) (*BaseCamera, error) {
	cam := &BaseCamera{screenWidth: width, screenHeight: height}
	// Init camera physics (required for cam movement)
	camBody := cp.NewKinematicBody()
	camBody.UserData = cam

	// Collision model
	cam.shape = cp.NewBox(camBody, float64(width), float64(height), 1)
	cam.shape.SetSensor(true)
	cam.shape.SetFilter(cp.SHAPE_FILTER_NONE)

	cam.ZoomFactor = camZoomFactorDefault

	return cam, nil
}

func (c *BaseCamera) IsVisible(entity GameEntity) bool {
	return c.bb().Intersects(entity.Shape().BB())
}

func (c *BaseCamera) VectorVisible(vec cp.Vector) bool {
	return c.bb().ContainsVect(vec)
}

// Returns center position of camera in world coordinates
func (c *BaseCamera) CenterWorldPos() cp.Vector {
	return c.ScreenToWorldPos(cp.Vector{float64(c.screenWidth) / 2, float64(c.screenHeight) / 2})
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

	// Offset by camera top left position
	res.Translate(-c.Position().X, -c.Position().Y)

	// We want to scale around center of image / screen
	// NOTE: Using UNSCALED viewport dimensions here on purpose
	res.Translate(-float64(c.screenWidth)*0.5, -float64(c.screenHeight)*0.5)
	res.Scale(
		c.ZoomFactor,
		c.ZoomFactor,
	)
	res.Translate(float64(c.screenWidth)*0.5, float64(c.screenHeight)*0.5)

	return res
}

// Draw filled rectangle on provided world position
func (c *BaseCamera) FillRect(topLeft, botRight cp.Vector, clr color.Color, antialias bool) {
	tLScreen := c.WorldToScreenPos(topLeft)
	bRScreen := c.WorldToScreenPos(botRight)
	diff := bRScreen.Sub(tLScreen)
	width := float32(diff.X)
	height := float32(diff.Y)
	vector.DrawFilledRect(c.screen, float32(tLScreen.X), float32(tLScreen.Y), width, height, clr, antialias)
}

// Draw stroked rectangle on provided world position
func (c *BaseCamera) StrokeRect(topLeft, botRight cp.Vector, strokeWidth float32, clr color.Color, antialias bool) {
	tLScreen := c.WorldToScreenPos(topLeft)
	bRScreen := c.WorldToScreenPos(botRight)
	diff := bRScreen.Sub(tLScreen)
	width := float32(diff.X)
	height := float32(diff.Y)
	vector.StrokeRect(c.screen, float32(tLScreen.X), float32(tLScreen.Y), width, height, strokeWidth, clr, antialias)
}

// NOTE: Radius will get scaled by camera zoom factor
func (c *BaseCamera) StrokeCircle(worldPos cp.Vector, r, strokeWidth float32, clr color.Color, antialias bool) {
	screenPos := c.WorldToScreenPos(worldPos)
	scaledRadius := r * float32(c.ZoomFactor)
	vector.StrokeCircle(c.screen, float32(screenPos.X), float32(screenPos.Y), scaledRadius, strokeWidth, clr, antialias)
}

func (c *BaseCamera) bb() cp.BB {
	// NOTE: Could deprecate. Top left pos = camera pos, but rather keep this in since its more robust to future changes
	topLeftWorldPosition, bottomRightWorldPosition := c.Viewport()
	return cp.BB{
		L: topLeftWorldPosition.X,
		B: topLeftWorldPosition.Y,
		R: bottomRightWorldPosition.X,
		T: bottomRightWorldPosition.Y,
	}
}

func (c *BaseCamera) Viewport() (cp.Vector, cp.Vector) {
	// NOTE: Could deprecate. Top left pos = camera pos, but rather keep this in since its more robust to future changes
	altTopLeft := c.ScreenToWorldPos(cp.Vector{0, 0})
	altBR := c.ScreenToWorldPos(cp.Vector{float64(c.screenWidth), float64(c.screenHeight)})
	return altTopLeft, altBR
}

func (c *BaseCamera) SetScreen(screen *ebiten.Image) { c.screen = screen }
func (c *BaseCamera) Zoom() float64                  { return c.ZoomFactor }

func (c *BaseCamera) DrawDebugInfo() {
	// Allow changing zoom with P-Up & P-Down
	if ebiten.IsKeyPressed(ebiten.KeyPageUp) && c.ZoomFactor < camZoomFactorMax {
		c.ZoomFactor += 0.005
	} else if ebiten.IsKeyPressed(ebiten.KeyPageDown) && c.ZoomFactor > camZoomFactorMin {
		c.ZoomFactor -= 0.005
	}

	screen := c.screen
	if screen == nil {
		log.Fatalln("Cannot draw debug info without screen")
		return
	}

	// Draw camera position & zoom
	ypos := 10
	tl, br := c.Viewport()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Viewport (world): (%.1f, %.1f) - (%.1f, %.1f)", tl.X, tl.Y, br.X, br.Y), 10, ypos)
	ypos += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Zoom %.3fx. Use P-UP & P-DOWN to adjust", c.ZoomFactor), 10, ypos)
	ypos += 15

	// Draw cursor position
	relX, relY := ebiten.CursorPosition()
	worldPos := c.ScreenToWorldPos(cp.Vector{float64(relX), float64(relY)})
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cursor Screen(%d, %d), World(%.1f, %.1f)", relX, relY, worldPos.X, worldPos.Y), 10, ypos)
}

func (c *BaseCamera) WorldToScreenPos(worldPos cp.Vector) cp.Vector {
	worldMatrix := c.worldMatrix()
	x, y := worldMatrix.Apply(worldPos.X, worldPos.Y)
	return cp.Vector{x, y}
}

func (c *BaseCamera) ScreenToWorldPos(screenPos cp.Vector) cp.Vector {
	worldMatrix := c.worldMatrix()
	if worldMatrix.IsInvertible() {
		worldMatrix.Invert()
		x, y := worldMatrix.Apply(screenPos.X, screenPos.Y)
		return cp.Vector{x, y}
	} else {
		// When scaling it can happened that matrix is not invertable
		log.Println("Error: could not calc world pos. World matrix not invertible")
		return screenPos
	}
}

func (c *BaseCamera) Position() cp.Vector   { return c.shape.Body().Position() }
func (c *BaseCamera) Body() *cp.Body        { return c.shape.Body() }
func (c *BaseCamera) Shape() *cp.Shape      { return c.shape }
func (c *BaseCamera) Screen() *ebiten.Image { return c.screen }
func (c *BaseCamera) ScreenWidth() int      { return c.screenWidth }
func (c *BaseCamera) ScreenHeight() int     { return c.screenHeight }
