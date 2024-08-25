package td

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/loot"
)

type TowerManager struct {
	world          engine.GameEntityManager
	goldManager    loot.GoldManager
	assetManager   engine.AssetManager
	worldMapReader engine.WorldMapReader

	touches map[ebiten.TouchID]time.Time
}

const (
	minDistanceBetweenTowers = float64(12.0)
	maxDistanceForDeletion   = float64(5.0)
	touchDurationForDeletion = float64(1.0)
	towerSizeX               = int(32)
	towerSizeY               = int(48)
	// TODO: Belongs to tower type
	costToBuy    = int64(50)
	refundIfSold = int64(50)
)

type TowerType int

const (
	SingleTarget TowerType = iota
	MultiTarget
)

var buildableTiles = []engine.MapTile{25, 26, 27, 28, 29, 31, 32, 33, 34, 37, 38, 39}

func NewTowerManager(
	world engine.GameEntityManager,
	am engine.AssetManager,
	goldManager loot.GoldManager,
	worldMapReader engine.WorldMapReader) (*TowerManager, error) {
	return &TowerManager{world: world, assetManager: am, goldManager: goldManager, worldMapReader: worldMapReader}, nil
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

func (t *TowerManager) Draw(screen *ebiten.Image) error {
	ebitenutil.DebugPrintAt(screen, "Use left mouse click to add towers", 20, 680)
	ebitenutil.DebugPrintAt(screen, "Use right mouse click to remove towers", 20, 700)
	return nil
}

func (t *TowerManager) AddTower(cursorPos cp.Vector) error {
	// Snap pos to 32x48 grid
	pos := engine.SnapToGrid(cursorPos, towerSizeX, towerSizeY)
	// Check if we're allowed to build on this map tile
	tile, err := t.worldMapReader.TileAt(pos)
	if err != nil {
		return fmt.Errorf("Unable to read tile data: %s", err.Error())
	}
	if !tileIsBuildable(tile) {
		return fmt.Errorf("Cannot build on tile %d", tile)
	}
	// Check if already occupied by other tower
	queryInfo := t.world.Space().PointQueryNearest(pos, minDistanceBetweenTowers, engine.TowerCollisionFilter())
	if queryInfo.Shape != nil {
		return fmt.Errorf("Collsion with existing tower")
	}
	// Check funds
	if !t.goldManager.CanAfford(costToBuy) {
		return fmt.Errorf("Insufficient funds!")
	}

	// Tower Factory
	// FIX:Currently type of tower is randomly selected
	selectedTower := TowerType(rand.Intn(2))
	var tower *TowerEntity
	switch selectedTower {
	case SingleTarget:
		tower, err = NewSingleTargetTower(t.world, t.assetManager)
	case MultiTarget:
		tower, err = NewMultiTargetTower(t.world, t.assetManager)
	default:
		err = fmt.Errorf("Invalid tower type provided")
	}
	if err != nil {
		return err
	}
	tower.shape.Body().SetPosition(pos)
	t.world.AddEntity(tower)
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
	t.goldManager.Refund(refundIfSold)
	return nil
}

func tileIsBuildable(tile engine.MapTile) bool {
	for _, iteratedTile := range buildableTiles {
		if iteratedTile == tile {
			return true
		}
	}
	return false
}
