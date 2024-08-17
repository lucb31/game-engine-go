package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type Camera interface {
	// Physics
	Position() cp.Vector
	Body() *cp.Body
	Shape() *cp.Shape

	// Drawing entities
	IsVisible(GameEntity) bool
	// Transforms world coordinates to camera coordinates
	ScreenPosition(cp.Vector) cp.Vector

	// General rendering
	DrawImage(*ebiten.Image, *ebiten.DrawImageOptions)
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

func (c *BaseCamera) IsVisible(entity GameEntity) bool {
	return false
	// a := c.Position()
	// b := c.Position().Add(cp.Vector{float64(c.viewportWidth), float64(c.viewportHeight)})
	// res := cp.ShapesCollide(c.Shape(), entity.Shape())
	// res := entity.Shape().BB().IntersectsSegment(a, b)
	// fmt.Println("checking visible", a, b, res, entity)
	// fmt.Println("shapes collide", res.Count > 0)
	// c.shape.Space().ShapeQuery(c.shape, func(shape *cp.Shape, points *cp.ContactPointSet) {
	// 	fmt.Println("Visible", shape)
	// })
	// return false
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

// ////////
// Getters
// ////////
func (c *BaseCamera) Position() cp.Vector { return c.shape.Body().Position() }
func (c *BaseCamera) Body() *cp.Body      { return c.shape.Body() }
func (c *BaseCamera) Shape() *cp.Shape    { return c.shape }
func (c *BaseCamera) ViewportWidth() int  { return c.viewportWidth }
func (c *BaseCamera) ViewportHeight() int { return c.viewportHeight }
