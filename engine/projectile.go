package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type CustomCollisionType uintptr

const (
	ProjectileCollision CustomCollisionType = iota
	PlayerCollision
)

type Projectile struct {
	id          GameEntityId
	shape       *cp.Shape
	Destination cp.Vector
	velocity    float64
	asset       *ProjectileAsset
	world       *GameWorld
}

type ProjectileAsset struct {
	Image *ebiten.Image
}

func NewProjectile(world *GameWorld, asset *ProjectileAsset, destination cp.Vector) (*Projectile, error) {
	if asset.Image == nil {
		return nil, fmt.Errorf("Failed to instantiate projectile. No asset provided")
	}
	p := &Projectile{world: world}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 200, Y: 200})
	body.SetVelocityUpdateFunc(p.calculateVelocity)
	body.UserData = p
	p.shape = cp.NewCircle(body, 8, cp.Vector{})
	p.shape.SetElasticity(0)
	p.shape.SetFriction(0)
	p.shape.SetCollisionType(cp.CollisionType(ProjectileCollision))
	p.Destination = destination
	p.velocity = 150
	p.asset = asset
	return p, nil
}

func (p *Projectile) Draw(screen *ebiten.Image) {
	// fmt.Println("Drawing projectile at", p.Shape.Body().Position())
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.shape.Body().Position().X, p.shape.Body().Position().Y)
	screen.DrawImage(p.asset.Image, &op)
}

func (p *Projectile) Id() GameEntityId      { return p.id }
func (p *Projectile) SetId(id GameEntityId) { p.id = id }
func (p *Projectile) Shape() *cp.Shape      { return p.shape }
func (p *Projectile) Destroy() {
	p.world.removeObject(p)
}

func (p *Projectile) calculateVelocity(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	position := body.Position()
	diff := p.Destination.Sub(position)
	diffNormalized := diff.Normalize()
	vel := diffNormalized.Mult(p.velocity)
	body.SetVelocityVector(vel)
}
