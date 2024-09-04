package engine

import (
	"fmt"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
	"github.com/lucb31/game-engine-go/engine/loot"
)

func ProjectileCollisionFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(cp.NO_GROUP, ProjectileCategory, NpcCategory|OuterWallsCategory)
}

type ProjectileTarget interface {
	// Needs physical body
	Body() *cp.Body
	// Needs damage model
	damage.Defender
}

type Projectile struct {
	*BaseEntityImpl

	// Physics
	shape          *cp.Shape
	velocity       float64
	direction      cp.Vector
	origin         cp.Vector
	staticVelocity cp.Vector

	// Logic
	// Gun this projectile was fired from
	gun      Gun
	target   ProjectileTarget
	piercing bool

	// Rendering
	asset *ProjectileAsset
}

type ProjectileAsset struct {
	Image          *ebiten.Image
	animationSpeed float64
	atp            AnimationTimeProvider
}

const defaultProjectileSpeed = float64(400.0)

func (a *ProjectileAsset) Draw(t RenderingTarget, position cp.Vector, angleInRad float64) error {
	op := ebiten.DrawImageOptions{}
	// Offset by half asset size to center position
	op.GeoM.Translate(-float64(a.Image.Bounds().Dx())/2, -float64(a.Image.Bounds().Dy())/2)
	// Add rotating animation
	if a.animationSpeed > 0.0 {
		animationFrameCount := 16
		animationFrame := int(a.atp.AnimationTime()/a.animationSpeed) % animationFrameCount
		op.GeoM.Rotate(2 * math.Pi / float64(animationFrameCount) * float64(animationFrame))
	} else {
		op.GeoM.Rotate(angleInRad)
	}

	// Translate to physical position
	op.GeoM.Translate(position.X, position.Y)
	t.DrawImage(a.Image, &op)
	return nil
}

func NewProjectile(gun Gun, asset *ProjectileAsset) (*Projectile, error) {
	if asset.Image == nil {
		return nil, fmt.Errorf("Failed to instantiate projectile. No asset provided")
	}
	base, err := NewBaseEntity()
	if err != nil {
		return nil, err
	}
	p := &Projectile{BaseEntityImpl: base, asset: asset}
	body := cp.NewKinematicBody()
	body.SetPosition(gun.Owner().Shape().Body().Position())
	body.SetVelocityUpdateFunc(p.calculateVelocity)
	body.UserData = p

	p.shape = cp.NewBox(body, 16, 16, 0)
	p.shape.SetSensor(true)
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

func (p *Projectile) Shape() *cp.Shape                  { return p.shape }
func (p *Projectile) Power() float64                    { return p.gun.Power() }
func (p *Projectile) AtkSpeed() float64                 { return 1.0 }
func (p *Projectile) LootTable() loot.LootTable         { return loot.NewEmptyLootTable() }
func (p *Projectile) SetTarget(target ProjectileTarget) { p.target = target }
func (p *Projectile) SetPiercing(piercing bool)         { p.piercing = piercing }

// Callback after projectile has hit an object. Used to implement projectile behaviour after hit
// Default: Remove projectile
// Piercing: Continue with current momentum
// Forking: TODO
func (p *Projectile) OnHit() error {
	if p.piercing {
		if p.target == nil {
			return fmt.Errorf("Whoops. No target. This should not happen")
		}
		p.staticVelocity = p.Shape().Body().Velocity()
		p.target = nil
		return nil
	}

	return p.Destroy()
}

func (p *Projectile) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	if p.target != nil {
		// Remove guided projectile if target no longer exists
		targetStillExists := p.shape.Space().ContainsBody(p.target.Body())
		if !targetStillExists {
			log.Println("Removing projectile: Target no longer exists")
			p.Destroy()
			return
		}

		p.direction = p.target.Body().Position()
	}

	// Remove projectile if fire range exceeded
	distanceFromOrigin := p.shape.Body().Position().Distance(p.origin)
	if math.IsNaN(distanceFromOrigin) || distanceFromOrigin >= p.gun.FireRange() {
		p.Destroy()
		return
	}

	// Move projectile by static velocity
	if p.staticVelocity.LengthSq() > 0.0 {
		body.SetVelocityVector(p.staticVelocity)
		return
	}

	// Move projectile towards destination position
	position := body.Position()
	diff := p.direction.Sub(position)
	diffNormalized := diff.Normalize()
	vel := diffNormalized.Mult(p.velocity)
	body.SetVelocityVector(vel)
}
