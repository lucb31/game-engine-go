package td

import (
	"fmt"
	"image/color"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type TowerEntity struct {
	*engine.BaseEntityImpl

	// Rendering
	asset     *engine.CharacterAsset
	animation string

	// Physics
	shape *cp.Shape

	// Logic
	gun engine.Gun
}

func NewTowerEntity(asset *engine.CharacterAsset) (*TowerEntity, error) {
	base, err := engine.NewBaseEntity()
	if err != nil {
		return nil, err
	}
	tower := &TowerEntity{BaseEntityImpl: base, asset: asset, animation: "idle"}
	body := cp.NewBody(1, cp.INFINITY)
	body.SetPosition(cp.Vector{X: 70, Y: 70})
	body.SetType(cp.BODY_KINEMATIC)
	body.SetPositionUpdateFunc(tower.Update)
	body.UserData = tower
	tower.shape = cp.NewBox(body, float64(towerSizeX), float64(towerSizeY), 0)
	tower.shape.SetFilter(engine.TowerCollisionFilter())
	return tower, nil
}

func NewSingleTargetTower(world engine.GameEntityManager, assetManager engine.AssetManager) (*TowerEntity, error) {
	// Init tower
	towerAsset, err := assetManager.CharacterAsset("tower-blue")
	if err != nil {
		return nil, fmt.Errorf("Could not find tower asset")
	}
	tower, err := NewTowerEntity(towerAsset)
	if err != nil {
		return nil, err
	}

	// Init gun
	projAsset, err := assetManager.ProjectileAsset("arrow")
	if err != nil {
		return nil, fmt.Errorf("Could not find projectile asset")
	}
	gunOpts := engine.BasicGunOpts{
		FireRatePerSecond: 1.5,
		FireRange:         250.0,
		Damage:            40,
	}
	tower.gun, err = engine.NewAutoAimGun(world, tower, projAsset, gunOpts)
	if err != nil {
		return nil, err
	}
	return tower, nil
}

func NewMultiTargetTower(world engine.GameEntityManager, assetManager engine.AssetManager) (*TowerEntity, error) {
	// Init tower
	towerAsset, err := assetManager.CharacterAsset("tower-red")
	if err != nil {
		return nil, fmt.Errorf("Could not find tower asset")
	}
	tower, err := NewTowerEntity(towerAsset)
	if err != nil {
		return nil, err
	}

	// Init gun
	projAsset, err := assetManager.ProjectileAsset("bone")
	if err != nil {
		return nil, fmt.Errorf("Could not find projectile asset")
	}
	gunOpts := engine.BasicGunOpts{
		FireRatePerSecond: 1.0,
		FireRange:         150.0,
		Damage:            25,
	}
	tower.gun, err = engine.NewShotGun(world, tower, projAsset, gunOpts)
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

func (t *TowerEntity) Draw(screen engine.RenderingTarget) error {
	t.DrawRange(screen)
	return t.asset.Draw(screen, t.shape, 0)
}

func (t *TowerEntity) DrawRange(screen engine.RenderingTarget) {
	screen.StrokeCircle(t.shape.Body().Position(), float32(t.gun.FireRange()), 2.0, color.RGBA{255, 0, 0, 0}, false)
}

func (n *TowerEntity) Shape() *cp.Shape          { return n.shape }
func (n *TowerEntity) LootTable() loot.LootTable { return loot.NewEmptyLootTable() }
