package engine

import (
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type GameEntityId int
type GameEntity interface {
	Id() GameEntityId
	SetId(GameEntityId)
	Shape() *cp.Shape
	Draw(RenderingTarget) error
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
