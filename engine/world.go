package engine

import (
	"fmt"
	"image/color"
	"log"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
	"github.com/lucb31/game-engine-go/engine/loot"
)

// ///////////////
// DEBUG SETTINGS
// ///////////////
const (
	DEBUG_CAMERA_POS             = true
	DEBUG_DRAW_STATIC_BODY       = false
	DEBUG_ENTITY_STATS           = true
	DEBUG_RENDER_COLLISION_BOXES = false
	SFX_VOLUME                   = 0.4
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

	WorldMap WorldMap
	FogOfWar FogOfWar
	Width    int64
	Height   int64
	// Integral of Physical time steps. Used for game sim
	gameTime *float64

	// Integral of time steps that continues after game over. Used for animation sim
	animationTime float64
	AssetManager  AssetManager
	space         *cp.Space

	// Game logic
	gameOver    bool
	GameSpeed   float64
	damageModel damage.DamageModel
}

func (w *GameWorld) drawVisibleObjects() {
	// Determine visible objects
	visibleObjectIds := []GameEntityId{}
	for id, obj := range w.objects {
		if w.EntityVisible(obj) {
			visibleObjectIds = append(visibleObjectIds, id)
		}
	}
	// Sort objects by id before drawing to ensure deterministic render order
	slices.SortFunc(visibleObjectIds, func(a, b GameEntityId) int {
		return int(a) - int(b)
	})
	for _, id := range visibleObjectIds {
		obj := w.objects[id]
		if err := obj.Draw(w.camera); err != nil {
			log.Printf("Error drawing object %d: %s \n", obj.Id(), err.Error())
		}
	}
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	if w.camera == nil {
		panic("Camera missing!")
	}
	w.camera.SetScreen(screen)
	w.WorldMap.Draw(w.camera)

	// Render player
	if w.player != nil {
		if err := w.player.Draw(w.camera); err != nil {
			log.Println("Error drawing player: ", err.Error())
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player pos: %s", w.player.shape.Body().Position()), 10, 90)
	}

	// Render entities that are visible in the camera viewport
	w.drawVisibleObjects()

	w.drawCombatLog()

	// Render fog of war
	if w.FogOfWar != nil {
		w.FogOfWar.Draw(w.camera)
	}

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
}

func (w *GameWorld) EntityVisible(e GameEntity) bool {
	// Skip rendering entities outside of camera viewport
	if !w.camera.IsVisible(e) {
		return false
	}
	// Skip rendering entities hidden in the fog of war
	if w.FogOfWar.VectorVisible(TopLeftBBPosition(e.Shape())) ||
		w.FogOfWar.VectorVisible(TopRightBBPosition(e.Shape())) ||
		w.FogOfWar.VectorVisible(BottomLeftBBPosition(e.Shape())) ||
		w.FogOfWar.VectorVisible(BottomRightBBPosition(e.Shape())) {
		return true
	}
	return false
}

func (w *GameWorld) Update() {
	dt := w.GameSpeed / 60.0
	w.animationTime += dt
	// Stop updating if game over
	if w.gameOver {
		return
	}
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
	object.SetEntityRemover(w)
	w.nextObjectId++
	return nil
}

// Removes an object from the world by scheduling for deletion
func (w *GameWorld) RemoveEntity(object BaseEntity) error {
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

func (w *GameWorld) AddCollisionLayer(mapData []byte) error {
	tileData, err := ReadCsvFromBinary(mapData)
	if err != nil {
		return err
	}

	// Register wall segments to physical space
	walls := CalcHorizontalWallSegments(tileData)
	walls = append(walls, CalcVerticalWallSegments(tileData)...)
	for _, wall := range walls {
		RegisterWallSegmentToSpace(w.space, wall)
	}
	return nil
}

// Adds a layer with collision segments AND tilesets
func (w *GameWorld) AddCombinedLayer(mapData []byte, tileset *Tileset) error {
	if err := w.AddLayer(mapData, tileset); err != nil {
		return err
	}
	return w.AddCollisionLayer(mapData)
}

func (w *GameWorld) AddLayer(mapData []byte, tileset *Tileset) error {
	return w.WorldMap.AddCsvLayer(mapData, tileset)
}

// Dropping items is done by evaluating the input loot table &
// spawning item sprites with guaranteed drops for every result of the original loot table
func (w *GameWorld) DropLoot(lootTable loot.LootTable, pos cp.Vector) error {
	for _, lootableItem := range lootTable.Result() {
		itemEntity, err := NewItemEntity(pos)
		if err != nil {
			return err
		}
		// FIX: Hard-coding wood asset here. This needs to be provided thorugh the loot table though
		asset, err := w.AssetManager.CharacterAsset("wood")
		if err != nil {
			return err
		}
		if err := itemEntity.SetAsset(asset); err != nil {
			return err
		}
		copyOfLootTable := loot.NewGuaranteedLootTable(lootableItem)
		itemEntity.SetLootTable(copyOfLootTable)
		if err := w.AddEntity(itemEntity); err != nil {
			return err
		}
	}
	return nil
}

func (w *GameWorld) EndGame()                        { w.gameOver = true }
func (w *GameWorld) Space() *cp.Space                { return w.space }
func (w *GameWorld) IngameTime() float64             { return *w.gameTime }
func (w *GameWorld) AnimationTime() float64          { return w.animationTime }
func (w *GameWorld) IsOver() bool                    { return w.gameOver }
func (w *GameWorld) DamageModel() damage.DamageModel { return w.damageModel }
func (w *GameWorld) Player() *Player                 { return w.player }

func (w *GameWorld) drawCombatLog() {
	damageLog := w.damageModel.DamageLog()
	entries := damageLog.Entries()
	for idx, entry := range entries {
		// Cleanup: Remove entries older than X seconds
		maxTimeDiff := 1.5
		timeDiff := w.IngameTime() - entry.GameTime
		if timeDiff > maxTimeDiff {
			if err := damageLog.RemoveByIdx(idx); err != nil {
				log.Fatalln("Could not remove log entry", entry, err.Error())
			}
			return
		}
		// Animate to scroll upwards
		absPos := entry.Pos.Add(cp.Vector{X: 0, Y: -timeDiff / maxTimeDiff * 20})
		relPos := w.camera.WorldToScreenPos(absPos)

		ebitenutil.DebugPrintAt(w.camera.Screen(), fmt.Sprintf("%.0f", entry.Damage), int(relPos.X), int(relPos.Y))
	}
}

// Actually remove a game entity from physics & object space
func (w *GameWorld) removeObject(id GameEntityId) {
	object, ok := w.objects[id]
	if !ok {
		log.Println("Oops, tried to delete unknown object", id)
		return
	}
	w.space.RemoveShape(object.Shape())
	w.space.RemoveBody(object.Shape().Body())
	delete(w.objects, id)
}

// Draw static bounding boxes for debugging purposes
func (w *GameWorld) drawDebugBoundingBoxes(screen *ebiten.Image) {
	w.Space().EachShape(func(shape *cp.Shape) {
		if shape.Body().GetType() == cp.BODY_STATIC {
			absStartPos := cp.Vector{shape.BB().L, shape.BB().B}
			relStartPos := w.camera.WorldToScreenPos(absStartPos)
			absEndPos := cp.Vector{shape.BB().R, shape.BB().T}
			relEndPos := w.camera.WorldToScreenPos(absEndPos)
			vector.StrokeLine(screen, float32(relStartPos.X), float32(relStartPos.Y), float32(relEndPos.X), float32(relEndPos.Y), 2.0, color.White, false)
		}
	})
}

// Debugging info for entities
func (w *GameWorld) drawEntityDebugInfo(screen *ebiten.Image) {
	yPos := 400
	visibleObjects := 0
	for _, obj := range w.objects {
		if w.EntityVisible(obj) {
			visibleObjects++
		}
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Visible Objects: (%d / %d)", visibleObjects, len(w.objects)), 10, yPos)
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
	yPos += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Shapes: %d", shapes), 10, yPos)
	yPos += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Projectiles: %d", projectiles), 10, yPos)
	yPos += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Npcs: %d", npcs), 10, yPos)
	yPos += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("# Fps: %0.1f", ebiten.ActualFPS()), 10, yPos)
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
	am, err := NewAssetManager(&w)
	if err != nil {
		return nil, err
	}
	w.AssetManager = am

	// Initialize walls on outer edge
	walls := []WallSegment{
		// Horizontal walls
		{cp.Vector{5, 5}, cp.Vector{float64(width) - 5, 5}},
		{cp.Vector{5, float64(height) - 5}, cp.Vector{float64(width) - 5, float64(height) - 5}},
		// Vertical walls
		{cp.Vector{5, 5}, cp.Vector{5, float64(height) - 5}},
		{cp.Vector{float64(width) - 5, 5}, cp.Vector{float64(width) - 5, float64(height) - 5}},
	}
	for _, wall := range walls {
		RegisterWallSegmentToSpace(space, wall)
	}

	return &w, nil
}

func NewGeneratedWorld(generator WorldGenerator) (*GameWorld, error) {
	// Init empty world
	width, height := generator.WorldDimensions()
	gameWorld, err := NewWorld(width, height)
	if err != nil {
		return nil, err
	}

	// Execute level generator
	res, err := generator.Generate(gameWorld.AssetManager)
	if err != nil {
		return nil, fmt.Errorf("Error during level generation: %s", err.Error())
	}

	// Apply map
	gameWorld.WorldMap = res.WorldMap

	// Generate objects
	for _, obj := range res.Objects {
		if err := gameWorld.AddEntity(obj); err != nil {
			return nil, fmt.Errorf("Error during object generation: %s", err.Error())
		}
	}

	return gameWorld, nil
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
}
