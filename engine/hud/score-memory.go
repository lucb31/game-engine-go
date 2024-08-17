package hud

import (
	"fmt"
	"slices"
)

type InMemoryScoreBoard struct {
	records []*ScoreRecord
}

func NewInMemoryScoreBoard() (*InMemoryScoreBoard, error) {
	fmt.Println("Initializing in mem scoreboard")
	return &InMemoryScoreBoard{}, nil
}

func (s *InMemoryScoreBoard) Save(val ScoreValue) error {
	rec, err := NewScoreRecord(val)
	if err != nil {
		return err
	}
	s.records = append(s.records, rec)
	// Sort by score DESC
	slices.SortFunc(s.records, func(a, b *ScoreRecord) int {
		return a.compareto(b)
	})
	return nil
}

func (s *InMemoryScoreBoard) IsHighscore(val ScoreValue) bool {
	highscore := s.Highscore()
	if highscore == nil {
		return true
	}
	return val >= highscore.Score
}

func (s *InMemoryScoreBoard) Highscore() *ScoreRecord {
	if len(s.records) == 0 {
		return nil
	}
	return s.records[0]
}

func (s *InMemoryScoreBoard) Print() error {
	fmt.Println("--- Memory Scoreboard ---")
	for idx, score := range s.records {
		fmt.Printf("[%d]: %s\n", idx, score.String())
	}
	return nil
}
