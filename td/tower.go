package td

import (
	"fmt"
	"time"

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
	shape *cp.Shape

	lastProjectileFired time.Time
}

const (
	towerFireRatePerSecond = float64(1.3)
)

func NewTower(world engine.GameEntityManager, asset *engine.CharacterAsset, projectile *engine.ProjectileAsset) (*TowerEntity, error) {
	tower := &TowerEntity{world: world, asset: asset, animation: "idle", projectileAsset: projectile}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 70, Y: 70})
	body.SetType(cp.BODY_KINEMATIC)
	body.SetPositionUpdateFunc(tower.Update)
	body.UserData = tower
	tower.shape = cp.NewBox(body, 32, 32, 0)
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
	now := time.Now()
	duration := float64(time.Second) / towerFireRatePerSecond
	if now.Sub(t.lastProjectileFired) < time.Duration(duration) {
		// Still reloading
		return
	}

	// Spawn projectile at tower position
	projectilePos := t.shape.Body().Position()
	proj, err := engine.NewProjectileWithDestination(t, t.world, t.projectileAsset, projectilePos, target.Shape().Body().Position())
	if err != nil {
		fmt.Println("Could not shoot projectile")
		return
	}
	t.world.AddEntity(proj)
	t.lastProjectileFired = time.Now()
}

func (t *TowerEntity) Draw(screen *ebiten.Image) {
	t.asset.Draw(screen, t.animation, t.shape.Body().Position())
}

func (t *TowerEntity) Destroy() {
	t.world.RemoveEntity(t)
}
func (n *TowerEntity) Id() engine.GameEntityId      { return n.id }
func (n *TowerEntity) SetId(id engine.GameEntityId) { n.id = id }
func (n *TowerEntity) Shape() *cp.Shape             { return n.shape }
