package td

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type TowerManager struct {
	world       engine.GameEntityManager
	goldManager engine.GoldManager

	towerAsset      *engine.CharacterAsset
	projectileAsset *engine.ProjectileAsset

	lastTowerSpawned time.Time
	touches          map[ebiten.TouchID]time.Time
}

const (
	minDistanceBetweenTowers = float64(12.0)
	maxDistanceForDeletion   = float64(5.0)
	touchDurationForDeletion = float64(1.0)
	costToBuy                = int64(50)
	refundIfSold             = int64(40)
)

func NewTowerManager(world engine.GameEntityManager, towerAsset *engine.CharacterAsset, projAsset *engine.ProjectileAsset, goldManager engine.GoldManager) (*TowerManager, error) {
	return &TowerManager{world: world, towerAsset: towerAsset, projectileAsset: projAsset, goldManager: goldManager}, nil
}

func (t *TowerManager) Update() {
	// Handle tower add / remove via touch
	newTouches := map[ebiten.TouchID]time.Time{}
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		pos := cp.Vector{float64(x), float64(y)}
		existingTouch, ok := t.touches[id]
		now := time.Now()
		if ok {
			// Remove tower on long touch
			newTouches[id] = existingTouch
			duration := float64(time.Second) * touchDurationForDeletion
			if now.Sub(existingTouch) < time.Duration(duration) {
				continue
			}
			if err := t.RemoveTower(pos); err != nil {
				fmt.Println("Could not remove tower: ", err.Error())
			}
		} else {
			// Add tower on new touch
			newTouches[id] = now
			if err := t.AddTower(pos); err != nil {
				fmt.Println("Could not add tower: ", err.Error())
			}
		}
	}
	t.touches = newTouches

	// Add tower on left-click
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if err := t.AddTower(cp.Vector{float64(mx), float64(my)}); err != nil {
			fmt.Println("Could not add tower: ", err.Error())
		}
	}

	// Remove tower on right-mouse click
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		err := t.RemoveTower(cp.Vector{float64(mx), float64(my)})
		if err != nil {
			fmt.Println("Could not remove tower: ", err.Error())
		}
	}

}

func (t *TowerManager) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "Use left mouse click to add towers", 20, 680)
	ebitenutil.DebugPrintAt(screen, "Use right mouse click to remove towers", 20, 700)
}

func (t *TowerManager) AddTower(pos cp.Vector) error {
	// Delay: Do not spawn more than 1 tower per second
	now := time.Now()
	duration := float64(time.Second) / 1
	if now.Sub(t.lastTowerSpawned) < time.Duration(duration) {
		return nil
	}
	// Avoid stacking towers
	// TODO: Tower grid to solve this
	queryInfo := t.world.Space().PointQueryNearest(pos, minDistanceBetweenTowers, engine.TowerCollisionFilter())
	if queryInfo.Shape != nil {
		return nil
	}
	// FIX: Avoid spawning towers when interacting with speed toggle
	if pos.Y >= 650 && pos.X >= 850 {
		return nil
	}
	// Check funds
	if !t.goldManager.CanAfford(costToBuy) {
		return fmt.Errorf("Insufficient funds!")
	}

	// Add tower entity
	tower, err := NewTower(t.world, t.towerAsset, t.projectileAsset)
	if err != nil {
		return err
	}
	tower.shape.Body().SetPosition(pos)
	t.world.AddEntity(tower)
	t.lastTowerSpawned = now
	// Spend gold
	t.goldManager.Remove(costToBuy)

	return nil
}

func (t *TowerManager) RemoveTower(pos cp.Vector) error {
	queryInfo := t.world.Space().PointQueryNearest(pos, maxDistanceForDeletion, engine.TowerCollisionFilter())
	if queryInfo.Shape == nil {
		return nil
	}
	tower, ok := queryInfo.Shape.Body().UserData.(*TowerEntity)
	if !ok {
		return fmt.Errorf("Collision checker did not return Tower Entity")
	}
	// Remove tower entity
	if err := tower.Destroy(); err != nil {
		return fmt.Errorf("Unable to remove tower: %s", err.Error())
	}
	// Refund gold
	t.goldManager.Add(refundIfSold)
	return nil
}
