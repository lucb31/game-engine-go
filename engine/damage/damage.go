package damage

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type DamageModel interface {
	ApplyDamage(atk Attacker, def Defender, gameTime float64) (*DamageRecord, error)
	DamageLog() DamageLog
}

type Attacker interface {
	Power() float64
}

type Defender interface {
	Health() float64
	Armor() float64
	Destroy() error
	SetHealth(float64)
	Shape() *cp.Shape
}

type BasicDamageModel struct {
	damageLog DamageLog
}

func NewBasicDamageModel() (*BasicDamageModel, error) {
	log, err := NewInMemoryDamageLog()
	if err != nil {
		return nil, err
	}
	return &BasicDamageModel{damageLog: log}, nil
}

func (m *BasicDamageModel) CalculateDamage(atk Attacker, def Defender) (float64, error) {
	// TODO: Percentage based armor
	damage := atk.Power() - def.Armor()
	// Fix to never return negative damage
	if damage < 0 {
		fmt.Println("Warning: Negative damage, this should not happen")
		return 0, nil
	}
	return damage, nil
}

func (m *BasicDamageModel) ApplyDamage(atk Attacker, def Defender, gameTime float64) (*DamageRecord, error) {
	damage, err := m.CalculateDamage(atk, def)
	if err != nil {
		return nil, err
	}
	newHealth := def.Health() - damage
	def.SetHealth(newHealth)
	rec := &DamageRecord{gameTime, damage, def.Shape().Body().Position()}
	err = m.damageLog.Add(*rec)
	if err != nil {
		fmt.Println("Could not log damage record. Continuing anyways...", *rec, err.Error())
	}
	if newHealth <= 0.0 {
		err := def.Destroy()
		if err != nil {
			return nil, err
		}
		return rec, nil
	}
	return rec, nil
}

func (m *BasicDamageModel) DamageLog() DamageLog { return m.damageLog }
