package engine

import "github.com/jakecoffman/cp"

type CreepProvider interface {
	NextNpc(wave Wave) (GameEntity, error)
}

type DefaultCreepProvider struct {
	asset *CharacterAsset
	opts  *NpcOpts
}

func NewDefaultCreepProvider(asset *CharacterAsset) (*DefaultCreepProvider, error) {
	opts := NpcOpts{
		Waypoints: []cp.Vector{
			{X: 48, Y: 720},
			{X: 976, Y: 720},
			{X: 976, Y: 48},
			{X: 208, Y: 48},
			{X: 208, Y: 560},
			{X: 816, Y: 560},
			{X: 816, Y: 208},
			{X: 368, Y: 208},
			{X: 368, Y: 384},
			{X: 640, Y: 384},
		},
	}

	return &DefaultCreepProvider{asset: asset, opts: &opts}, nil
}

func (p *DefaultCreepProvider) NextNpc(wave Wave) (GameEntity, error) {
	npc, err := NewNpc(p.asset, *p.opts)
	if err != nil {
		return nil, err
	}
	return npc, nil
}
