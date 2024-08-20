package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

// ///////////////
// DEBUG SETTINGS
// ///////////////
const (
	DEBUG_CAMERA_POS             = false
	DEBUG_DRAW_STATIC_BODY       = false
	DEBUG_ENTITY_STATS           = true
	DEBUG_RENDER_COLLISION_BOXES = true
)

type GameWorld struct {
	// Entity management
	objects      map[GameEntityId]GameEntity
	player       *Player
	nextObjectId GameEntityId
	// Removing object from the world needs to be buffered towards the end of a timestep
	objectIdsToDelete []GameEntityId

	// Rendering
	camera Camera

	WorldMap *WorldMap
	Width    int64
	Height   int64
	// Number of frames drawn. Used for animation
	FrameCount int64
	// Integral of Physical time steps. Used for game sim
	gameTime     *float64
	AssetManager AssetManager
	space        *cp.Space

	// Game logic
	gameOver    bool
	GameSpeed   float64
	damageModel damage.DamageModel
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	if w.camera == nil {
		panic("Camera missing!")
	}
	w.camera.SetScreen(screen)
	w.WorldMap.Draw(w.camera)
	// Debugging options
	if DEBUG_CAMERA_POS {
		w.camera.DrawDebugInfo()
	}
	if DEBUG_DRAW_STATIC_BODY {
		w.drawDebugBoundingBoxes(screen)
	}
	if DEBUG_ENTITY_STATS {
		w.drawEntityDebugInfo(screen)
	}
	if w.gameOver {
		return
	}

	// Render player
	if w.player != nil {
		if err := w.player.Draw(w.camera); err != nil {
			fmt.Println("Error drawing player: ", err.Error())
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player pos: %s", w.player.shape.Body().Position()), 10, 90)
	}
	// Render entities that are visible in the camera viewport
	for _, obj := range w.objects {
		if w.camera.IsVisible(obj) {
			if err := obj.Draw(w.camera); err != nil {
				fmt.Printf("Error drawing object %d: %s \n", obj.Id(), err.Error())
			}
		}
	}

	w.drawCombatLog()
}

func (w *GameWorld) Update() {
	// Stop updating if game over
	if w.gameOver {
		return
	}
	w.FrameCount++
	dt := w.GameSpeed / 60.0
	*w.gameTime += dt
	w.space.Step(dt)
	// Delete objects scheduled for deletion
	if len(w.objectIdsToDelete) > 0 {
		for _, id := range w.objectIdsToDelete {
			w.removeObject(id)
		}
		w.objectIdsToDelete = []GameEntityId{}
	}
}

// Adds a game entity to the world by
// - registering to physics space
// - registering in object map to add / find / remove entities
func (w *GameWorld) AddEntity(object GameEntity) error {
	w.space.AddBody(object.Shape().Body())
	w.space.AddShape(object.Shape())
	w.objects[w.nextObjectId] = object
	object.SetId(w.nextObjectId)
	w.nextObjectId++
	return nil
}

// Removes an object from the world by scheduling for deletion
func (w *GameWorld) RemoveEntity(object GameEntity) error {
	idToDelete := object.Id()
	// Check for duplicates
	for _, id := range w.objectIdsToDelete {
		if id == idToDelete {
			return fmt.Errorf("Already scheduled for deletion")
		}
	}
	w.objectIdsToDelete = append(w.objectIdsToDelete, idToDelete)
	return nil
}

func (w *GameWorld) AddCollisionLayer(mapData []byte, tileset *Tileset) error {
	if err := w.WorldMap.AddLayer(mapData, tileset); err != nil {
		return err
	}

	// Register wall segments to physical space
	tileData := w.WorldMap.layers[len(w.WorldMap.layers)-1].TileData()
	walls := CalcHorizontalWallSegments(tileData)
	walls = append(walls, CalcVerticalWallSegments(tileData)...)
	for _, wall := range walls {
		RegisterWallSegmentToSpace(w.space, wall)
	}
	return nil
}

func (w *GameWorld) GetEntities() *map[GameEntityId]GameEntity { return &w.objects }
func (w *GameWorld) EndGame()                                  { w.gameOver = true }
func (w *GameWorld) Space() *cp.Space                          { return w.space }
func (w *GameWorld) GetIngameTime() float64                    { return *w.gameTime }
func (w *GameWorld) IsOver() bool                              { return w.gameOver }
func (w *GameWorld) DamageModel() damage.DamageModel           { return w.damageModel }

func (w *GameWorld) drawCombatLog() {
	log := w.damageModel.DamageLog()
	entries := log.Entries()
	for idx, entry := range entries {
		// Cleanup: Remove entries older than X seconds
		maxTimeDiff := 1.5
		timeDiff := w.GetIngameTime() - entry.GameTime
		if timeDiff > maxTimeDiff {
			if err := log.RemoveByIdx(idx); err != nil {
				fmt.Println("Could not remove log entry", entry, err.Error())
			}
			return
		}
		// Animate to scroll upwards
		absPos := entry.Pos.Add(cp.Vector{X: 0, Y: -timeDiff / maxTimeDiff * 20})
		relPos := w.camera.AbsToRel(absPos)

		ebitenutil.DebugPrintAt(w.camera.Screen(), fmt.Sprintf("%.0f", entry.Damage), int(relPos.X), int(relPos.Y))
	}
}

// Actually remove a game entity from physics & object space
func (w *GameWorld) removeObject(id GameEntityId) {
	object, ok := w.objects[id]
	if !ok {
		fmt.Println("Oops, tried to delete unknown object", id)
		return
	}
	w.space.RemoveShape(object.Shape())
	w.space.RemoveBody(object.Shape().Body())
	delete(w.objects, id)
}

func (w *GameWorld) initializeOuterWallBoundingBox(camera Camera) {
	minX := float64(camera.ViewportWidth() / 2)
	minY := float64(camera.ViewportHeight() / 2)
	maxX := float64(w.Width) - minX
	maxY := float64(w.Height) - minY

	walls := []WallSegment{
		{cp.Vector{minX, minY}, cp.Vector{minX, maxY}},
		{cp.Vector{maxX, minY}, cp.Vector{maxX, maxY}},
		{cp.Vector{minX, minY}, cp.Vector{maxX, minY}},
		{cp.Vector{minX, maxY}, cp.Vector{maxX, maxY}},
	}
	for _, wall := range walls {
		RegisterWallSegmentToSpace(w.space, wall)
	}
}

// Draw static bounding boxes for debugging purposes
func (w *GameWorld) drawDebugBoundingBoxes(screen *ebiten.Image) {
	w.Space().EachShape(func(shape *cp.Shape) {
		if shape.Body().GetType() == cp.BODY_STATIC {
			absStartPos := cp.Vector{shape.BB().L, shape.BB().B}
			relStartPos := w.camera.AbsToRel(absStartPos)
			absEndPos := cp.Vector{shape.BB().R, shape.BB().T}
			relEndPos := w.camera.AbsToRel(absEndPos)
			vector.StrokeLine(screen, float32(relStartPos.X), float32(relStartPos.Y), float32(relEndPos.X), float32(relEndPos.Y), 2.0, color.White, false)
		}
	})
}

// Debugging info for entities
func (w *GameWorld) drawEntityDebugInfo(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Objects: %d", len(w.objects)), 10, 30)
	shapes := 0
	projectiles := 0
	npcs := 0
	w.Space().EachShape(func(s *cp.Shape) {
		if _, is := s.Body().UserData.(*Projectile); is {
			projectiles++
		} else if _, is := s.Body().UserData.(*NpcEntity); is {
			npcs++
		}
		shapes++
	})
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Shapes: %d", shapes), 10, 45)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Projectiles: %d", projectiles), 10, 60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Npcs: %d", npcs), 10, 75)
}

func BoundingBoxFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, OuterWallsCategory, PlayerCategory|NpcCategory|TowerCategory|ProjectileCategory)
}

func NewWorld(width int64, height int64) (*GameWorld, error) {
	// Intialize damage model
	damageModel, err := damage.NewBasicDamageModel()
	if err != nil {
		return nil, err
	}

	// Initialize physics
	gameTime := float64(0)
	space, err := NewPhysicsSpace(damageModel, &gameTime)
	if err != nil {
		return nil, err
	}
	w := GameWorld{
		gameTime:    &gameTime,
		Width:       width,
		Height:      height,
		space:       space,
		damageModel: damageModel,
		objects:     map[GameEntityId]GameEntity{},
		GameSpeed:   1.0,
	}
	// Initialize assets
	am, err := NewAssetManager(&w.FrameCount)
	if err != nil {
		return nil, err
	}
	w.AssetManager = am

	return &w, nil
}

func (w *GameWorld) InitPlayer(am AssetManager) (*Player, error) {
	// Initialize player (after world has been initialized to reference it)
	playerAsset, err := am.CharacterAsset("ranger")
	if err != nil {
		return nil, err
	}
	projAsset, err := am.ProjectileAsset("arrow")
	if err != nil {
		return nil, err
	}
	player, err := NewPlayer(w, playerAsset, projAsset)
	if err != nil {
		return nil, err
	}
	// Explicitly NOT adding the player to the object space via addObject.
	// Might want to revisit this later
	w.space.AddBody(player.Shape().Body())
	w.space.AddShape(player.Shape())
	w.player = player
	return player, nil
}

func (w *GameWorld) SetCamera(camera Camera) {
	w.camera = camera
	w.space.AddBody(w.camera.Body())
	w.space.AddShape(w.camera.Shape())

	w.initializeOuterWallBoundingBox(camera)
}
