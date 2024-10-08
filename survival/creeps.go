package survival

import (
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine"
)

type NpcType struct {
	assetName string
	opts      engine.NpcOpts
}

// ////////
// CONFIG
// ////////

// TODO: Config file or DB
var availableNpcs = []NpcType{
	{"npc-torch", engine.NpcOpts{BasePower: 30, BaseHealth: 60, BaseMovementSpeed: 75.0, GoldValue: 2}},
	{"npc-orc", engine.NpcOpts{BasePower: 15, BaseHealth: 90, BaseMovementSpeed: 50.0, GoldValue: 3}},
	{"npc-slime", engine.NpcOpts{BasePower: 50, BaseHealth: 30, BaseMovementSpeed: 25.0, GoldValue: 1}},
}

type SurvCreepProvider struct {
	assetManager engine.AssetManager
	target       engine.DefenderEntity
	// Required to not spawn within current viewport
	camera engine.Camera
	// Spawn areas
	spawnAreaLayer engine.MapLayer
	// Map waypoints used to unstuck creeps
	aiWaypoints *engine.WaypointInfo
}

func NewSurvCreepProvider(am engine.AssetManager, t engine.DefenderEntity, cam engine.Camera) (*SurvCreepProvider, error) {
	return &SurvCreepProvider{assetManager: am, target: t, camera: cam}, nil
}

func (p *SurvCreepProvider) ParseNoSpawnArea(width, height int64, mapCsvData []byte) error {
	layer, err := engine.NewBaseMapLayer(width, height, mapCsvData, nil)
	if err != nil {
		return err
	}
	p.spawnAreaLayer = layer
	return nil
}

// NOTE: Space is required to calculate graph based on WP distances and collision between
func (p *SurvCreepProvider) ParseCreepWaypoints(mapCsvData []byte, space *cp.Space) error {
	mapTiles, err := engine.ReadCsvFromBinary(mapCsvData)
	if err != nil {
		return err
	}
	// Determine wp positions from CSV data
	wpPositions := []cp.Vector{}
	for row := range len(mapTiles) {
		for col := range len(mapTiles[row]) {
			if mapTiles[row][col] == engine.EmptyTile {
				continue
			}
			x, y := engine.GridPosToCenterWorldPos(col, row)
			wpPositions = append(wpPositions, cp.Vector{x, y})
		}
	}

	// Build dijkstra graph for pathfinding based on wp positions
	p.aiWaypoints, err = engine.NewWaypointInfo(space, wpPositions)
	if err != nil {
		return err
	}

	return nil
}

func (p *SurvCreepProvider) NextNpc(wave engine.Wave) (engine.GameEntity, error) {
	npcType := p.nextNpcType()
	// Load asset
	npcAsset, err := p.assetManager.CharacterAsset(npcType.assetName)
	if err != nil {
		return nil, err
	}
	// Load opts & calculate starting position
	opts := npcType.opts
	opts.WaypointInfo = *p.aiWaypoints
	opts.StartingPos, err = p.calcCreepSpawnPosition(p.camera)
	if err != nil {
		return nil, err
	}
	// Apply scaling
	opts.BaseHealth = wave.HealthScalingFunc(opts.BaseHealth)

	// Init npc
	npc, err := engine.NewNpcAggro(p.target, npcAsset, opts)
	if err != nil {
		return nil, err
	}
	return npc, nil
}

// Creeps spawning is restricted by
// - map layer that determines spawnable areas
// - camera viewport
func (p *SurvCreepProvider) calcCreepSpawnPosition(cam engine.Camera) (cp.Vector, error) {
	for tries := 0; tries < 10; tries++ {
		// Select random position within spawnable area
		width, height := p.spawnAreaLayer.Dimensions()
		row := rand.IntN(height)
		col := rand.IntN(width)
		randX, randY := engine.GridPosToCenterWorldPos(col, row)

		// Check spawn area layer
		tileAt, err := p.spawnAreaLayer.TileAt(cp.Vector{randX, randY})
		if err != nil {
			log.Printf("Error checking spawnable area. Retrying...\n")
			continue
		}
		if tileAt == -1 {
			log.Printf("Not a spawnable area (%d). Retrying...\n", tileAt)
			continue
		}

		// Check if within camera viewport
		if cam.VectorVisible(cp.Vector{randX, randY}) {
			log.Println("Position within viewport. Retrying...", randX, randY)
			continue
		}
		return cp.Vector{X: randX, Y: randY}, nil
	}
	return cp.Vector{}, fmt.Errorf("Could not find a spawn position. Max tries reached")
}

// Choose a random npc type to spawn next
func (p *SurvCreepProvider) nextNpcType() NpcType {
	idx := rand.IntN(len(availableNpcs))
	return availableNpcs[idx]
}
