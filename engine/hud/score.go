package hud

import (
	"fmt"
	"time"
)

type ScoreValue float64
type ScoreBoard interface {
	// Saves the score record to the score board
	Save(ScoreValue) error
	// Returns true if provided score record would result in a new highscore
	IsHighscore(ScoreValue) bool
	// Retrieve current highscore details
	Highscore() *ScoreRecord
	// Print scoreboard to console
	Print() error
}

type ScoreRecord struct {
	Timestamp int64
	Score     ScoreValue
}

func NewScoreRecord(val ScoreValue) (*ScoreRecord, error) {
	return &ScoreRecord{time.Now().Unix(), val}, nil
}

func (r *ScoreRecord) StringSlice() []string {
	res := []string{fmt.Sprint(r.Timestamp), fmt.Sprint(r.Score)}
	return res
}

// Used in sort functions
func (a *ScoreRecord) compareto(b *ScoreRecord) int {
	if a.Score > b.Score {
		return -1
	} else if a.Score < b.Score {
		return 1
	} else {
		return 0
	}
}

func (r *ScoreRecord) String() string {
	time := time.Unix(r.Timestamp, 0)
	return fmt.Sprintf("%f (%s)", r.Score, time)
}
