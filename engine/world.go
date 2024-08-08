package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

// TODO: Add debug util to increase this via keybinds
const gameSpeed = float64(1.0)

type GameWorld struct {
	objects      map[GameEntityId]GameEntity
	player       GameEntity
	WorldMap     *WorldMap
	Width        int64
	Height       int64
	FrameCount   int64
	AssetManager *AssetManager
	Space        *cp.Space

	nextObjectId GameEntityId
	// Removing object from the world needs to be buffered towards the end of a timestep
	objectIdsToDelete []GameEntityId
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	w.WorldMap.Draw(screen)
	// TODO: Currently drawing ALL objects. Fine as long as there is no camera movement
	for _, obj := range w.objects {
		obj.Draw(screen)
	}
	w.player.Draw(screen)
}

func (w *GameWorld) Update() {
	w.FrameCount++
	w.Space.Step(1.0 / 60.0 * gameSpeed)
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
	w.Space.AddBody(object.Shape().Body())
	w.Space.AddShape(object.Shape())
	w.objects[w.nextObjectId] = object
	object.SetId(w.nextObjectId)
	w.nextObjectId++
	return nil
}

// Removes an object from the world by scheduling for deletion
func (w *GameWorld) RemoveEntity(object GameEntity) error {
	w.objectIdsToDelete = append(w.objectIdsToDelete, object.Id())
	return nil
}

func (w *GameWorld) GetEntities() *map[GameEntityId]GameEntity {
	return &w.objects
}

// Actually remove a game entity from physics & object space
func (w *GameWorld) removeObject(id GameEntityId) {
	object, ok := w.objects[id]
	if !ok {
		fmt.Println("Oops, tried to delete unknown object", id)
		return
	}
	w.Space.RemoveShape(object.Shape())
	w.Space.RemoveBody(object.Shape().Body())
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
	return cp.NewShapeFilter(0, uint(OuterWallsCategory), uint(PlayerCategory|NpcCategory|TowerCategory|ProjectileCategory))
}

func NewWorld(width int64, height int64) (*GameWorld, error) {
	// Initialize physics
	space, err := NewPhysicsSpace()
	if err != nil {
		return nil, err
	}
	initializeBoundingBox(space, float64(width), float64(height))
	w := GameWorld{
		Width:   width,
		Height:  height,
		Space:   space,
		objects: map[GameEntityId]GameEntity{},
	}
	// Initialize assets
	am, err := NewAssetManager(&w.FrameCount)
	if err != nil {
		return nil, err
	}
	w.AssetManager = am

	// Initialize player (after world has been initialized to reference it)
	asset, ok := am.CharacterAssets["player"]
	if !ok {
		return nil, fmt.Errorf("Could not find player asset")
	}
	projAsset, ok := am.ProjectileAssets["bone"]
	if !ok {
		return nil, fmt.Errorf("Could not find projectile asset")
	}
	player, err := NewPlayer(&w, &asset, &projAsset)
	if err != nil {
		return &w, err
	}
	w.player = player
	// Explicitly NOT adding the player to the object space via addObject.
	// Might want to revisit this later
	space.AddBody(player.Shape().Body())
	space.AddShape(player.Shape())

	return &w, nil
}
