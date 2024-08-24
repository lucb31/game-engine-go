package engine

import (
	"math"

	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type GameEntityId int
type GameEntity interface {
	Id() GameEntityId
	SetId(GameEntityId)
	Shape() *cp.Shape
	Draw(RenderingTarget) error
	Lootable
}

// Interface for an entity that can provide loot
type Lootable interface {
	LootTable() *LootTable
	Destroy() error
}

type EntityRemover interface {
	RemoveEntity(object GameEntity) error
}

type IngameTimeProvider interface {
	GetIngameTime() float64
}

type GameEntityManager interface {
	EntityRemover
	IngameTimeProvider
	AddEntity(object GameEntity) error
	GetEntities() *map[GameEntityId](GameEntity)
	Space() *cp.Space
	DamageModel() damage.DamageModel
	EndGame()
}

type GameEntityStatReader interface {
	Armor() float64
	AtkSpeed() float64
	Health() float64
	MaxHealth() float64
	Power() float64
	MovementSpeed() float64
}

type GameEntityStatWriter interface {
	SetArmor(v float64)
	SetAtkSpeed(v float64)
	SetHealth(h float64)
	SetPower(v float64)
	SetMaxHealth(v float64)
	SetMovementSpeed(v float64)
}

type GameEntityStatReadWriter interface {
	GameEntityStatReader
	GameEntityStatWriter
}

type GameEntityStats struct {
	armor         float64
	atkSpeed      float64
	health        float64
	maxHealth     float64
	movementSpeed float64
	power         float64
}

func DefaultGameEntityStats() GameEntityStats {
	return GameEntityStats{0, 1.0, 100.0, 100.0, 100.0, 30.0}
}
func (s *GameEntityStats) Armor() float64         { return s.armor }
func (s *GameEntityStats) AtkSpeed() float64      { return s.atkSpeed }
func (s *GameEntityStats) Health() float64        { return s.health }
func (s *GameEntityStats) MaxHealth() float64     { return s.maxHealth }
func (s *GameEntityStats) Power() float64         { return s.power }
func (s *GameEntityStats) MovementSpeed() float64 { return s.movementSpeed }

func (s *GameEntityStats) SetArmor(v float64)         { s.armor = v }
func (s *GameEntityStats) SetAtkSpeed(v float64)      { s.atkSpeed = v }
func (s *GameEntityStats) SetHealth(h float64)        { s.health = math.Min(h, s.maxHealth) }
func (s *GameEntityStats) SetPower(v float64)         { s.power = v }
func (s *GameEntityStats) SetMaxHealth(v float64)     { s.maxHealth = v }
func (s *GameEntityStats) SetMovementSpeed(v float64) { s.movementSpeed = v }
