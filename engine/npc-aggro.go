package engine

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

// NPC AI that will move towards target entity once it enters aggro range
type NpcAggro struct {
	*NpcEntity

	target GameEntity
	// True, once target has entered aggro range
	engaged bool
}

func NewNpcAggro(remover EntityRemover, target GameEntity, asset *CharacterAsset, opts NpcOpts) (*NpcAggro, error) {
	if target == nil {
		return nil, fmt.Errorf("Did not receive target")
	}
	base, err := NewNpc(remover, asset, opts)
	if err != nil {
		return nil, err
	}
	npc := &NpcAggro{NpcEntity: base, target: target}
	npc.Shape().Body().SetVelocityUpdateFunc(npc.velocityByAggro)
	return npc, nil
}

func (n *NpcAggro) velocityByAggro(body *cp.Body, gravity cp.Vector, damping float64, dt float64) {
	// Check if already engaged
	if n.engaged {
		n.moveTowards(body, n.target.Shape().Body().Position())
		return
	}
	n.engaged = true

	return
}
