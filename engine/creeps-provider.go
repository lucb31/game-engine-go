package engine

import "github.com/jakecoffman/cp"

type CreepProvider interface {
	NextNpc(remover EntityRemover, opts NpcOpts) (GameEntity, error)
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

func (p *DefaultCreepProvider) NextNpc(remover EntityRemover, opts NpcOpts) (GameEntity, error) {
	// TODO: Not optimal. Better would be a "merge options" utility method here
	opts.Waypoints = p.opts.Waypoints
	npc, err := NewNpc(remover, p.asset, opts)
	if err != nil {
		return nil, err
	}
	return npc, nil
}
