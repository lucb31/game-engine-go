package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type GameWorld struct {
	objects  map[GameEntityId]GameEntity
	player   GameEntity
	WorldMap *WorldMap
	Width    int64
	Height   int64
	// Number of frames drawn. Used for animation
	FrameCount int64
	// Integral of Physical time steps. Used for game sim
	gameTime     float64
	AssetManager *AssetManager
	space        *cp.Space

	nextObjectId GameEntityId
	// Removing object from the world needs to be buffered towards the end of a timestep
	objectIdsToDelete []GameEntityId

	gameOver  bool
	GameSpeed float64
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	w.WorldMap.Draw(screen)
	if w.gameOver {
		return
	}
	// TODO: Currently drawing ALL objects. Fine as long as there is no camera movement
	for _, obj := range w.objects {
		obj.Draw(screen)
	}
	if w.player != nil {
		w.player.Draw(screen)
	}
}

func (w *GameWorld) Update() {
	// Stop updating if game over
	if w.gameOver {
		return
	}
	w.FrameCount++
	dt := w.GameSpeed / 60.0
	w.gameTime += dt
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

func (w *GameWorld) GetEntities() *map[GameEntityId]GameEntity {
	return &w.objects
}

func (w *GameWorld) EndGame() {
	w.gameOver = true
}

func (w *GameWorld) Space() *cp.Space {
	return w.space
}

func (w *GameWorld) GetIngameTime() float64 {
	return w.gameTime
}

func (w *GameWorld) IsOver() bool {
	return w.gameOver
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

func createFences(am *AssetManager, space *cp.Space) ([]GameEntity, error) {
	fenceData := [][]int{
		{-1, -1, -1, -1, -1, -1, -1, -1},
		{-1, -1, -1, -1, 1, 14, 14, 3},
		{-1, -1, -1, -1, 4, -1, -1, 4},
		{-1, -1, -1, -1, 4, -1, -1, 4},
		{-1, -1, -1, -1, 9, 14, 14, 10},
	}
	objects := []GameEntity{}
	// Temporary disable fence
	return objects, nil
	for row, rowData := range fenceData {
		for col, tileIdx := range rowData {
			if tileIdx > -1 {
				im, err := am.GetTile("fences", tileIdx)
				if err != nil {
					return nil, err
				}
				body := cp.NewStaticBody()
				body.SetPosition(cp.Vector{X: float64(mapTileSize * col), Y: float64(mapTileSize * row)})
				shape := cp.NewBox(body, 16, 16, 0)
				//shape := cp.NewCircle(body, 8, cp.Vector{})

				shape.SetFriction(1)
				shape.SetElasticity(0)
				space.AddBody(body)
				space.AddShape(shape)
				objects = append(objects, &StaticGameEntity{shape: shape, Image: im})
			}
		}
	}
	return objects, nil
}

func initializeBoundingBox(space *cp.Space, width float64, height float64) {
	offset := 0.0
	walls := []cp.Vector{
		{offset, offset}, {offset, height},
		{width, offset}, {width, height},
		{offset, offset}, {width, offset},
		{offset, height}, {width, height},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 2))
		shape.SetElasticity(1)
		shape.SetFriction(1)
		shape.SetFilter(BoundingBoxFilter())
	}
}

func BoundingBoxFilter() cp.ShapeFilter {
	return cp.NewShapeFilter(0, OuterWallsCategory, PlayerCategory|NpcCategory|TowerCategory|ProjectileCategory)
}

func NewWorld(width int64, height int64) (*GameWorld, error) {
	// Initialize physics
	space, err := NewPhysicsSpace()
	if err != nil {
		return nil, err
	}
	initializeBoundingBox(space, float64(width), float64(height))
	w := GameWorld{
		Width:     width,
		Height:    height,
		space:     space,
		objects:   map[GameEntityId]GameEntity{},
		GameSpeed: 1.0,
	}
	// Initialize assets
	am, err := NewAssetManager(&w.FrameCount)
	if err != nil {
		return nil, err
	}
	w.AssetManager = am

	return &w, nil
}

func initPlayer(w *GameWorld, am *AssetManager) error {
	// Initialize player (after world has been initialized to reference it)
	asset, ok := am.CharacterAssets["player"]
	if !ok {
		return fmt.Errorf("Could not find player asset")
	}
	projAsset, ok := am.ProjectileAssets["bone"]
	if !ok {
		return fmt.Errorf("Could not find projectile asset")
	}
	player, err := NewPlayer(w, &asset, &projAsset)
	if err != nil {
		return err
	}
	// Explicitly NOT adding the player to the object space via addObject.
	// Might want to revisit this later
	w.space.AddBody(player.Shape().Body())
	w.space.AddShape(player.Shape())
	w.player = player
	return nil
}
