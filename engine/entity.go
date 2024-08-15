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

type GameEntityManager interface {
	AddEntity(object GameEntity) error
	RemoveEntity(object GameEntity) error
	GetEntities() *map[GameEntityId](GameEntity)
	GetIngameTime() float64
	Space() *cp.Space
	EndGame()
}
