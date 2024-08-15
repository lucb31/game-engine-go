package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
)

type GameEntityId int
type GameEntity interface {
	Id() GameEntityId
	SetId(GameEntityId)
	Shape() *cp.Shape
	Draw(*ebiten.Image)
	Destroy() error
}

type EntityRemover interface {
	RemoveEntity(object GameEntity) error
}

type GameEntityManager interface {
	EntityRemover
	AddEntity(object GameEntity) error
	GetEntities() *map[GameEntityId](GameEntity)
	GetIngameTime() float64
	Space() *cp.Space
	EndGame()
}
