package engine

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

func ProjectileCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(cp.NO_GROUP, ProjectileCategory, NpcCategory|OuterWallsCategory)
}

type Projectile struct {
	// Entity management
	id    GameEntityId
	world GameEntityManager

	// Physics
	shape    *cp.Shape
	velocity float64

	// Logic
	// Gun this projectile was fired from
	gun       Gun
	target    GameEntity
	direction cp.Vector
	origin    cp.Vector

	// Rendering
	asset *ProjectileAsset
}

type ProjectileAsset struct {
	Image          *ebiten.Image
	currentFrame   *int64
	animationSpeed int
}

const defaultProjectileSpeed = float64(300.0)

func (a *ProjectileAsset) Draw(t RenderingTarget, position cp.Vector, angleInRad float64) error {
	op := ebiten.DrawImageOptions{}
	// Offset by half asset size to center position
	op.GeoM.Translate(-float64(a.Image.Bounds().Dx())/2, -float64(a.Image.Bounds().Dy())/2)
	// Add rotating animation
	if a.animationSpeed > 0 {
		animationFrameCount := 16
		animationFrame := int(*a.currentFrame/int64(a.animationSpeed)) % animationFrameCount
		op.GeoM.Rotate(2 * math.Pi / float64(animationFrameCount) * float64(animationFrame))
	} else {
		op.GeoM.Rotate(angleInRad)
	}

	// Translate to physical position
	op.GeoM.Translate(position.X, position.Y)
	t.DrawImage(a.Image, &op)
	return nil
}

func NewProjectile(gun Gun, world GameEntityManager, asset *ProjectileAsset) (*Projectile, error) {
	if asset.Image == nil {
		return nil, fmt.Errorf("Failed to instantiate projectile. No asset provided")
	}
	p := &Projectile{world: world, asset: asset}
	body := cp.NewKinematicBody()
	body.SetPosition(gun.Owner().Shape().Body().Position())
	body.SetVelocityUpdateFunc(p.calculateVelocity)
	body.UserData = p

	p.shape = cp.NewBox(body, 16, 16, 0)
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(ProjectileCollision))
	p.shape.SetFilter(ProjectileCollisionFilter())
	p.velocity = defaultProjectileSpeed
	p.gun = gun
	p.origin = body.Position()
	return p, nil
}

func (p *Projectile) Draw(t RenderingTarget) error {
	angle := p.Shape().Body().Position().Sub(p.direction).Neg().ToAngle()
	return p.asset.Draw(t, p.shape.Body().Position(), angle)
}

func (p *Projectile) Id() GameEntityId      { return p.id }
func (p *Projectile) SetId(id GameEntityId) { p.id = id }
func (p *Projectile) Shape() *cp.Shape      { return p.shape }
func (p *Projectile) Power() float64        { return p.gun.Power() }
func (p *Projectile) Destroy() error {
	return p.world.RemoveEntity(p)
}

func (p *Projectile) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Remove guided projectile if target no longer exists
	if p.target != nil {
		// TODO: Utilize physics space query to find target. Then we can remove the whole "GetEntities" idea
		_, ok := (*p.world.GetEntities())[p.target.Id()]
		if !ok {
			p.Destroy()
			return
		}

		p.direction = p.target.Shape().Body().Position()
	}

	// Remove projectile if fire range exceeded
	distanceTravelled := p.shape.Body().Position().Distance(p.origin)
	if math.IsNaN(distanceTravelled) || distanceTravelled >= p.gun.FireRange() {
		p.Destroy()
		return
	}

	// Remove projectile if destination reached
	position := body.Position()
	diff := p.direction.Sub(position)
	if diff.Length() < 0.1 {
		p.Destroy()
		return
	}

	diffNormalized := diff.Normalize()
	vel := diffNormalized.Mult(p.velocity)
	body.SetVelocityVector(vel)
}
