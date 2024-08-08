package engine

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type CustomCollisionType cp.CollisionType

const (
	PlayerCollision CustomCollisionType = iota
	ProjectileCollision
	NpcCollision
)

type CollisionCategory uint

const (
	PlayerCategory CollisionCategory = iota + 1
	NpcCategory
	OuterWallsCategory
	InnerWallsCategory
	TowerCategory
	ProjectileCategory
)

func ProjectileCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, uint(ProjectileCategory), uint(NpcCategory|OuterWallsCategory&^PlayerCategory))
}

type Projectile struct {
	// Entity management
	id    GameEntityId
	world GameEntityManager

	// Physics
	shape    *cp.Shape
	velocity float64

	// Logic
	owner       GameEntity
	Destination cp.Vector

	// Rendering
	asset *ProjectileAsset
}

type ProjectileAsset struct {
	Image          *ebiten.Image
	currentFrame   *int64
	animationSpeed int
}

func (a *ProjectileAsset) Draw(screen *ebiten.Image, position cp.Vector) error {
	op := ebiten.DrawImageOptions{}
	// Offset by half asset size to center position
	op.GeoM.Translate(-float64(a.Image.Bounds().Dx())/2, -float64(a.Image.Bounds().Dy())/2)
	// Add rotating animation
	animationFrameCount := 16
	animationFrame := int(*a.currentFrame/int64(a.animationSpeed)) % animationFrameCount
	op.GeoM.Rotate(2 * math.Pi / float64(animationFrameCount) * float64(animationFrame))
	// Translate to physical position
	op.GeoM.Translate(position.X, position.Y)
	screen.DrawImage(a.Image, &op)
	return nil
}

func NewProjectileWithDestination(owner GameEntity, world GameEntityManager, asset *ProjectileAsset, startPosition cp.Vector, endPosition cp.Vector) (*Projectile, error) {
	if asset.Image == nil {
		return nil, fmt.Errorf("Failed to instantiate projectile. No asset provided")
	}
	p := &Projectile{world: world}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(startPosition)
	body.SetVelocityUpdateFunc(p.calculateVelocity)
	body.UserData = p
	p.shape = cp.NewBox(body, 16, 16, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(ProjectileCollision))
	p.shape.SetFilter(ProjectileCollisionFilter())
	p.velocity = 300
	p.asset = asset
	p.Destination = endPosition
	p.owner = owner
	return p, nil
}

func NewProjectileWithOrientation(owner GameEntity, world GameEntityManager, asset *ProjectileAsset, position cp.Vector, orientation Orientation) (*Projectile, error) {
	destination := destinationFromOrientation(orientation)
	return NewProjectileWithDestination(owner, world, asset, position, destination)
}

func destinationFromOrientation(orientation Orientation) cp.Vector {
	switch orientation {
	case North:
		return cp.Vector{0, -1000}
	case South:
		return cp.Vector{0, 1000}
	case East:
		return cp.Vector{1000, 0}
	default:
		return cp.Vector{-1000, 0}
	}
}

func (p *Projectile) Draw(screen *ebiten.Image) {
	p.asset.Draw(screen, p.shape.Body().Position())
}

func (p *Projectile) Id() GameEntityId      { return p.id }
func (p *Projectile) SetId(id GameEntityId) { p.id = id }
func (p *Projectile) Shape() *cp.Shape      { return p.shape }
func (p *Projectile) Destroy() {
	p.world.RemoveEntity(p)
}

func (p *Projectile) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	position := body.Position()
	diff := p.Destination.Sub(position)
	// Remove projectile if destination reached
	if diff.Length() < 5 {
		p.Destroy()
		return
	}
	diffNormalized := diff.Normalize()
	vel := diffNormalized.Mult(p.velocity)
	body.SetVelocityVector(vel)
}
