package damage

import (
	"fmt"

	"github.com/jakecoffman/cp"
)

type DamageRecord struct {
	GameTime float64
	Damage   float64
	Pos      cp.Vector
	Fatal    bool
}

type DamageLog interface {
	Add(rec DamageRecord) error
	Entries() []DamageRecord
	RemoveByIdx(idx int) error
}

type InMemoryDamageLog struct {
	entries []DamageRecord
}

func NewInMemoryDamageLog() (*InMemoryDamageLog, error) {
	return &InMemoryDamageLog{}, nil
}

func (l *InMemoryDamageLog) Add(rec DamageRecord) error {
	l.entries = append(l.entries, rec)
	return nil
}

func (l *InMemoryDamageLog) Entries() []DamageRecord { return l.entries }
func (l *InMemoryDamageLog) RemoveByIdx(idx int) error {
	if idx < 0 || idx > len(l.entries)-1 {
		return fmt.Errorf("Out of bounds")
	}
	l.entries = append(l.entries[:idx], l.entries[idx+1:]...)
	return nil
}
