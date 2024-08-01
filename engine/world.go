package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type Biome int
type GameEntityId int

const (
	Gras        Biome = 32
	Rock        Biome = 42
	Undef       Biome = 71
	mapTileSize       = 16
)

type GameEntity interface {
	Id() GameEntityId
	SetId(GameEntityId)
	Shape() *cp.Shape
	Draw(*ebiten.Image)
	Destroy()
}

type GameWorld struct {
	objects      map[GameEntityId]GameEntity
	player       GameEntity
	Biome        [][]Biome
	Width        int64
	Height       int64
	FrameCount   int64
	AssetManager *AssetManager
	space        *cp.Space
	nextObjectId GameEntityId

	// Removing object from the world needs to be buffered towards the end of a timestep
	objectIdsToDelete []GameEntityId
}

func (w *GameWorld) Draw(screen *ebiten.Image) {
	w.drawBiomes(screen)
	// TODO: Currently drawing ALL objects. Fine as long as there is no camera movement
	for _, obj := range w.objects {
		obj.Draw(screen)
	}
	w.player.Draw(screen)
}

func (w *GameWorld) Update() {
	w.FrameCount++
	w.space.Step(1.0 / 60.0)
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
func (w *GameWorld) addObject(object GameEntity) {
	fmt.Println("Adding object", object)
	w.space.AddBody(object.Shape().Body())
	w.space.AddShape(object.Shape())
	w.objects[w.nextObjectId] = object
	object.SetId(w.nextObjectId)
	w.nextObjectId++
}

// Removes an object from the world by scheduling for deletion
func (w *GameWorld) DestroyObject(object GameEntity) {
	w.objectIdsToDelete = append(w.objectIdsToDelete, object.Id())
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

func (w *GameWorld) drawBiomes(screen *ebiten.Image) {
	// Drawing WHOLE map. This is ok because there is no camera movement right now
	for row := range w.Height {
		for col := range w.Width {
			// Set tile position
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(col*mapTileSize), float64(row*mapTileSize))

			// Select correct tile from tileset
			subIm, err := w.AssetManager.GetTile("plains", int(w.Biome[row][col]))
			if err != nil {
				fmt.Println("Unable to draw biome cell", err.Error())
				return
			}
			screen.DrawImage(subIm, &op)
		}
	}
}

func createBiome(width, height int64) ([][]Biome, error) {
	// TODO: Should be coming from file / external source
	mapData := [][]Biome{
		{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9},
		{13, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 15},
	}
	biome := make([][]Biome, height)
	// Copy map data in & fill remaining cells with placeholder tile
	for row := range height {
		biome[row] = make([]Biome, width)
		for col := range width {
			if int64(len(mapData)) > row && int64(len(mapData[row])) > col {
				biome[row][col] = mapData[row][col]
			} else {
				biome[row][col] = Undef
			}
		}
	}
	return biome, nil
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

func NewPhysicsSpace() (*cp.Space, error) {
	// Initialize physics
	space := cp.NewSpace()
	// Initialize bounding box
	offset := 0.0
	walls := []cp.Vector{
		{offset, offset}, {offset, 240},
		{320, offset}, {320, 240},
		{offset, offset}, {320, offset},
		{offset, 240}, {320, 240},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 2))
		shape.SetElasticity(1)
		shape.SetFriction(1)
	}
	// Register collision handlers
	handler := space.NewCollisionHandler(cp.CollisionType(ProjectileCollision), cp.CollisionType(NpcCollision))
	handler.BeginFunc = npcProjectilecollisionHandler
	space.NewWildcardCollisionHandler(cp.CollisionType(ProjectileCollision)).PostSolveFunc = removeProjectile
	return space, nil
}

func removeProjectile(arb *cp.Arbiter, space *cp.Space, userData interface{}) {
	p, _ := arb.Bodies()

	projectile, ok := p.UserData.(*Projectile)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return
	}
	projectile.Destroy()
}

func npcProjectilecollisionHandler(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
	// Validate correct collision partners & type assert
	a, b := arb.Bodies()
	projectile, ok := a.UserData.(*Projectile)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Projectile")
		return false
	}
	npc, ok := b.UserData.(*NpcEntity)
	if !ok {
		fmt.Println("Type assertion for projectile collision failed. Did not receive valid Npc", b.UserData)
		return false
	}
	// Trigger projectile hit with COPY of projectile
	npc.OnProjectileHit(*projectile)
	// Remove projectile
	projectile.Destroy()
	return false
}

func NewWorld(width int64, height int64) (*GameWorld, error) {
	// Initialize assets
	am, err := NewAssetManager()
	if err != nil {
		return nil, err
	}
	// Initialize map
	biome, err := createBiome(width, height)
	if err != nil {
		return nil, err
	}
	// Initialize physics
	space, err := NewPhysicsSpace()
	if err != nil {
		return nil, err
	}
	// Initialize some fences
	// objects, err := createFences(am, space)
	// if err != nil {
	// 	return nil, err
	// }
	w := GameWorld{
		Biome:        biome,
		Width:        width,
		Height:       height,
		AssetManager: am,
		space:        space,
		objects:      map[GameEntityId]GameEntity{},
	}

	// Initialize player (after world has been initialized to reference it)
	asset, ok := am.CharacterAssets["player"]
	if !ok {
		return nil, fmt.Errorf("Could not find player asset")
	}
	projAsset, ok := am.ProjectileAssets["bone"]
	if !ok {
		return nil, fmt.Errorf("Could not find projectile asset")
	}
	// TODO: Find a better / generic solution to give assets access to the current frame count
	asset.currentFrame = &w.FrameCount
	player, err := NewPlayer(&w, &asset, &projAsset)
	if err != nil {
		return &w, err
	}
	w.player = player
	// Explicitly NOT adding the player to the object space via addObject.
	// Might want to revisit this later
	space.AddBody(player.Shape().Body())
	space.AddShape(player.Shape())

	// Initialize an npc
	npc, err := NewNpc(&w, &asset)
	if err != nil {
		return &w, err
	}
	w.addObject(npc)
	return &w, nil
}
