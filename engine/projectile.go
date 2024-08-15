package engine

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
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
	owner     GameEntity
	target    GameEntity
	direction cp.Vector

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

func NewProjectileWithTarget(owner GameEntity, target GameEntity, world GameEntityManager, asset *ProjectileAsset) (*Projectile, error) {
	p, err := newProjectile(owner, world, asset)
	if err != nil {
		return nil, err
	}
	p.target = target
	return p, nil
}

func newProjectile(owner GameEntity, world GameEntityManager, asset *ProjectileAsset) (*Projectile, error) {
	if asset.Image == nil {
		return nil, fmt.Errorf("Failed to instantiate projectile. No asset provided")
	}
	p := &Projectile{world: world, asset: asset}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(owner.Shape().Body().Position())
	body.SetVelocityUpdateFunc(p.calculateVelocity)
	body.UserData = p
	p.shape = cp.NewBox(body, 16, 16, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(ProjectileCollision))
	p.shape.SetFilter(ProjectileCollisionFilter())
	p.velocity = 300
	p.owner = owner
	return p, nil
}

func NewProjectileWithDirection(owner GameEntity, world GameEntityManager, asset *ProjectileAsset, endPosition cp.Vector) (*Projectile, error) {
	p, err := newProjectile(owner, world, asset)
	if err != nil {
		return nil, err
	}
	p.direction = endPosition
	return p, nil
}

func NewProjectileWithOrientation(owner GameEntity, world GameEntityManager, asset *ProjectileAsset, orientation Orientation) (*Projectile, error) {
	destination := directionFromOrientationAndPos(orientation, owner.Shape().Body().Position())
	return NewProjectileWithDirection(owner, world, asset, destination)
}

func directionFromOrientationAndPos(orientation Orientation, pos cp.Vector) cp.Vector {
	switch orientation {
	case North:
		return cp.Vector{pos.X, -1000}
	case South:
		return cp.Vector{pos.X, 1000}
	case East:
		return cp.Vector{1000, pos.Y}
	default:
		return cp.Vector{-1000, pos.Y}
	}
}

func (p *Projectile) Draw(screen *ebiten.Image) {
	p.asset.Draw(screen, p.shape.Body().Position())
}

func (p *Projectile) Id() GameEntityId      { return p.id }
func (p *Projectile) SetId(id GameEntityId) { p.id = id }
func (p *Projectile) Shape() *cp.Shape      { return p.shape }
func (p *Projectile) Destroy() error {
	return p.world.RemoveEntity(p)
}

func (p *Projectile) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	direction := p.direction
	if p.target != nil {
		direction = p.target.Shape().Body().Position()
	}
	position := body.Position()
	diff := direction.Sub(position)
	diffNormalized := diff.Normalize()
	vel := diffNormalized.Mult(p.velocity)
	body.SetVelocityVector(vel)
}
