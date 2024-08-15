package td

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type TowerEntity struct {
	id    engine.GameEntityId
	world engine.GameEntityManager

	// Rendering
	asset           *engine.CharacterAsset
	projectileAsset *engine.ProjectileAsset
	animation       string

	// Physics
	shape               *cp.Shape
	lastProjectileFired float64
}

const (
	towerFireRatePerSecond = float64(1.5)
)

func NewTower(world engine.GameEntityManager, asset *engine.CharacterAsset, projectile *engine.ProjectileAsset) (*TowerEntity, error) {
	tower := &TowerEntity{world: world, asset: asset, animation: "idle", projectileAsset: projectile}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 70, Y: 70})
	body.SetType(cp.BODY_KINEMATIC)
	body.SetPositionUpdateFunc(tower.Update)
	body.UserData = tower
	tower.shape = cp.NewBox(body, 32, 32, 0)
	tower.shape.SetFilter(engine.TowerCollisionFilter())
	return tower, nil
}

func (t *TowerEntity) Update(body *cp.Body, dt float64) {
	var target engine.GameEntity
	for _, val := range *t.world.GetEntities() {
		npc, ok := val.(*engine.NpcEntity)
		if ok {
			target = npc
			break
		}
	}
	t.shoot(target)
}

func (t *TowerEntity) shoot(target engine.GameEntity) {
	if target == nil {
		return
	}
	// Timeout until reloaded
	now := t.world.GetIngameTime()
	diff := now - t.lastProjectileFired
	if diff < 1/towerFireRatePerSecond {
		return
	}

	// Spawn projectile
	proj, err := engine.NewProjectileWithTarget(t, target, t.world, t.projectileAsset)
	if err != nil {
		fmt.Println("Could not shoot projectile")
		return
	}
	t.world.AddEntity(proj)
	t.lastProjectileFired = now
}

func (t *TowerEntity) Draw(screen *ebiten.Image) {
	t.asset.DrawRectBoundingBox(screen, t.shape)
	t.asset.Draw(screen, t.animation, t.shape.Body().Position())
}

func (t *TowerEntity) Destroy() error {
	return t.world.RemoveEntity(t)
}
func (n *TowerEntity) Id() engine.GameEntityId      { return n.id }
func (n *TowerEntity) SetId(id engine.GameEntityId) { n.id = id }
func (n *TowerEntity) Shape() *cp.Shape             { return n.shape }
