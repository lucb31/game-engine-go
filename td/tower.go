package td

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type TowerEntity struct {
	id    engine.GameEntityId
	world engine.GameEntityManager

	// Rendering
	asset     *engine.CharacterAsset
	animation string

	// Physics
	shape *cp.Shape

	// Logic
	gun engine.Gun
}

const (
	towerFireRatePerSecond = float64(1.5)
)

func NewTower(world engine.GameEntityManager, asset *engine.CharacterAsset, projectile *engine.ProjectileAsset) (*TowerEntity, error) {
	tower := &TowerEntity{world: world, asset: asset, animation: "idle"}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 70, Y: 70})
	body.SetType(cp.BODY_KINEMATIC)
	body.SetPositionUpdateFunc(tower.Update)
	body.UserData = tower
	tower.shape = cp.NewBox(body, float64(towerSizeX), float64(towerSizeY), 0)
	tower.shape.SetFilter(engine.TowerCollisionFilter())

	var err error
	gunOpts := engine.BasicGunOpts{
		FireRatePerSecond: towerFireRatePerSecond,
		FireRange:         tower.towerRange(),
	}
	tower.gun, err = engine.NewAutoAimGun(world, tower, projectile, gunOpts)
	if err != nil {
		return nil, err
	}
	return tower, nil
}

func (t *TowerEntity) Update(body *cp.Body, dt float64) {
	if t.gun.IsReloading() {
		return
	}
	t.gun.Shoot()
}

func (t *TowerEntity) towerRange() float64 {
	return 250.0
}

func (t *TowerEntity) Draw(screen *ebiten.Image) {
	t.asset.Draw(screen, t.animation, t.shape)
	t.DrawRange(screen)
}

func (t *TowerEntity) DrawRange(screen *ebiten.Image) {
	vector.StrokeCircle(screen, float32(t.shape.Body().Position().X), float32(t.shape.Body().Position().Y), float32(t.towerRange()), 2.0, color.RGBA{255, 0, 0, 0}, false)
}

func (t *TowerEntity) Destroy() error {
	return t.world.RemoveEntity(t)
}
func (n *TowerEntity) Id() engine.GameEntityId      { return n.id }
func (n *TowerEntity) SetId(id engine.GameEntityId) { n.id = id }
func (n *TowerEntity) Shape() *cp.Shape             { return n.shape }
