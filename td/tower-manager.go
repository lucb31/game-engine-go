package td

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type TowerManager struct {
	world engine.GameEntityManager

	towerAsset      *engine.CharacterAsset
	projectileAsset *engine.ProjectileAsset

	lastTowerSpawned time.Time
}

const minDistanceBetweenTowers = float64(12.0)

func NewTowerManager(world engine.GameEntityManager, towerAsset *engine.CharacterAsset, projAsset *engine.ProjectileAsset) (*TowerManager, error) {
	return &TowerManager{world: world, towerAsset: towerAsset, projectileAsset: projAsset}, nil
}

func (t *TowerManager) Update() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		t.AddTower(cp.Vector{float64(mx), float64(my)})
	}
}

func (t *TowerManager) Draw(screen *ebiten.Image) {
}

// Initialize a tower
func (t *TowerManager) AddTower(pos cp.Vector) error {
	// Avoid stacking towers
	// TODO: Tower grid to solve this
	queryInfo := t.world.Space().PointQueryNearest(pos, minDistanceBetweenTowers, engine.TowerCollisionFilter())
	if queryInfo.Shape != nil {
		return nil
	}
	// Delay: Do not spawn more than 1 tower per second
	now := time.Now()
	duration := float64(time.Second) / 1
	if now.Sub(t.lastTowerSpawned) < time.Duration(duration) {
		return nil
	}

	fmt.Println("Spawning tower at", pos)
	tower, err := NewTower(t.world, t.towerAsset, t.projectileAsset)
	if err != nil {
		return err
	}
	tower.shape.Body().SetPosition(pos)
	t.world.AddEntity(tower)
	t.lastTowerSpawned = now

	return nil
}
