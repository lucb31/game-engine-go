package engine

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/jakecoffman/cp"
	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine/damage"
)

type HarvestingTool interface {
	damage.Attacker
	// Returns true if there is a harvestable entity within range
	InRange() bool
	// Returns true if currently harvesting
	Harvesting() bool
	// Returns nil if nothing in range
	Nearest() Harvestable
	// Needs to be called by owner every tick to ensure swing timers are working correctly
	Update() error

	// Start harvesting nearest harvestable object
	HarvestNearest() error
	Abort() error

	// Dependency management
	SetAnimationController(AnimationController)
}

type Harvestable interface {
	damage.Defender
	BaseEntity
	Lootable
	Position() cp.Vector
}

var HarvestToolCollisionFilter = cp.NewShapeFilter(cp.NO_GROUP, cp.ALL_CATEGORIES, HarvestableCategory)
var HarvestableCollisionFilter = cp.NewShapeFilter(0, HarvestableCategory, PlayerCategory)

const (
	defaultHarvestingSpeed = 1.5
	defaultHarvestingRange = 50.0
	defaultHarvestingPower = 50.0
	// Delay in seconds between starting to harvest and playing harvest sfx.
	// Required to properly sync sfx with animation & harvesting speed
	harvestingSfxDelay = 0.45
)

type WoodHarvestingTool struct {
	// Dependencies
	owner               GameEntity
	em                  GameEntityManager
	animationController AnimationController

	// Stats
	// Maximum distance to harvestable
	harvestingRange float64
	// Multiplicator for harvest swing timer
	harvestingSpeed float64
	harvestingPower float64

	// State
	swingTimer Timeout
	target     Harvestable
	harvesting bool

	// SFX
	player *audio.Player
	// Used to sync harvesting SE with animation
	sfxTimeout Timeout
}

func NewWoodHarvestingTool(em GameEntityManager, owner GameEntity) (*WoodHarvestingTool, error) {
	if owner == nil {
		return nil, fmt.Errorf("Cannot init wood harvesting tool without owner")
	}
	ht := &WoodHarvestingTool{owner: owner, em: em}

	// Setup timers
	var err error
	if ht.swingTimer, err = NewIngameTimeout(em); err != nil {
		return nil, err
	}
	if ht.sfxTimeout, err = NewIngameTimeout(em); err != nil {
		return nil, err
	}

	// Set defaults
	ht.harvestingRange = defaultHarvestingRange
	ht.harvestingSpeed = defaultHarvestingSpeed
	ht.harvestingPower = defaultHarvestingPower
	return ht, nil
}

func (ht *WoodHarvestingTool) Nearest() Harvestable {
	query := ht.owner.Shape().Space().PointQueryNearest(ht.owner.Shape().Body().Position(), ht.harvestingRange, HarvestToolCollisionFilter)
	if query.Shape == nil {
		return nil
	}
	harvestable, ok := query.Shape.Body().UserData.(Harvestable)
	if !ok {
		log.Println("Expected harvestable target, but found something else", query.Shape.Body().UserData)
		return nil
	}
	return harvestable
}

func (ht *WoodHarvestingTool) InRange() bool     { return ht.Nearest() != nil }
func (ht *WoodHarvestingTool) Harvesting() bool  { return ht.harvesting }
func (ht *WoodHarvestingTool) AtkSpeed() float64 { return ht.harvestingSpeed }
func (ht *WoodHarvestingTool) Power() float64    { return ht.harvestingPower }
func (ht *WoodHarvestingTool) SetAnimationController(ac AnimationController) {
	ht.animationController = ac
}

func (ht *WoodHarvestingTool) PlayHarvestingSE() error {
	// Init player first time, else rewind
	if ht.player == nil {
		reader := bytes.NewReader(assets.PunchTreeOGG)
		stream, err := vorbis.DecodeWithoutResampling(reader)
		if err != nil {
			return err
		}
		ht.player, err = audio.CurrentContext().NewPlayer(stream)
		if err != nil {
			return err
		}
		ht.player.SetVolume(SFX_VOLUME)
	} else if err := ht.player.Rewind(); err != nil {
		return err
	}
	ht.player.Play()
	return nil
}

func (ht *WoodHarvestingTool) Update() error {
	if !ht.harvesting {
		return nil
	}
	// Check if we need to play harvesting SFX
	if ht.sfxTimeout.Done() {
		if err := ht.PlayHarvestingSE(); err != nil {
			return fmt.Errorf("Could not play harvesting sound effect: %s", err.Error())
		}
		// Disable timeout until next swing
		ht.sfxTimeout.Stop()
	}
	// Check if we can swing again
	if !ht.swingTimer.Done() {
		return nil
	}

	// HARVESTING SWING
	// Apply damage via damage model
	spaceUserData, ok := ht.owner.Shape().Space().StaticBody.UserData.(SpaceUserData)
	if !ok {
		return fmt.Errorf("Cannot apply harvest damage: No damage model found\n")
	}
	rec, err := spaceUserData.damageModel.ApplyDamage(ht, ht.target, ht.em.IngameTime())
	if err != nil {
		return fmt.Errorf("Could not apply damage: %e\n", err)
	}
	// Fatal -> Harvesting done
	if rec.Fatal {
		log.Println("Finished harvesting")
		// Disabling. Currently damage model is destroying
		// if err := ht.target.Destroy(); err != nil {
		// 	return err
		// }

		// Drop loot at slightly randomized position
		pos := ht.target.Position().Add(cp.Vector{rand.Float64()*20 - 10, rand.Float64()*20 - 10})
		if err := ht.em.DropLoot(ht.target.LootTable(), pos); err != nil {
			return err
		}

		// Directly add loot for harvesting target
		// if err := ht.owner.Inventory().AddLoot(ht.target.LootTable()); err != nil {
		// 	return err
		// }

		// Reset harvesting tool
		return ht.Abort()
	}

	// If not fatal, queue up the next swing, animation & sfx
	ht.swingTimer.Set(1 / ht.AtkSpeed())
	if err := ht.animationController.Play("harvest"); err != nil {
		return err
	}
	ht.sfxTimeout.Set(harvestingSfxDelay)
	return nil
}

func (ht *WoodHarvestingTool) HarvestNearest() error {
	target := ht.Nearest()
	if target == nil {
		return fmt.Errorf("Nothin in range")
	}
	// TODO: Throw error. Ensure we are not spamming harvest nearest
	if ht.harvesting {
		return nil
	}
	log.Println("Starting harvest", target)
	ht.target = target
	ht.harvesting = true
	ht.swingTimer.Set(1 / ht.AtkSpeed())
	// Play animation & queue sfx
	if err := ht.animationController.Play("harvest"); err != nil {
		return err
	}
	ht.sfxTimeout.Set(harvestingSfxDelay)
	return nil
}

func (ht *WoodHarvestingTool) Abort() error {
	if !ht.harvesting {
		return nil
	}
	log.Println("aborting harvest")
	ht.harvesting = false
	ht.target = nil
	ht.swingTimer.Stop()
	ht.sfxTimeout.Stop()
	ht.animationController.StopPlaying()
	if ht.player != nil {
		ht.player.Pause()
	}
	return nil
}
