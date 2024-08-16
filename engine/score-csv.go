package engine

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
)

type CsvScoreBoard struct {
	path string
}

func NewCsvScoreKeeper(path string) (*CsvScoreBoard, error) {
	fmt.Println("Initializing csv scoreboard")
	return &CsvScoreBoard{path}, nil
}

func (c *CsvScoreBoard) Save(score ScoreValue) error {
	f, err := os.OpenFile(c.path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	rec, err := NewScoreRecord(score)
	if err != nil {
		return err
	}
	csvWriter := csv.NewWriter(f)
	if err := csvWriter.Write(rec.StringSlice()); err != nil {
		return err
	}
	csvWriter.Flush()
	return nil
}

func (c *CsvScoreBoard) IsHighscore(score ScoreValue) bool {
	highscore := c.Highscore()
	if highscore == nil {
		return false
	}
	return score >= highscore.Score
}

func (c *CsvScoreBoard) Highscore() *ScoreRecord {
	_, err := c.readRecordsFromCsv()
	if err != nil {
		fmt.Println("Could not retrieve highscore: ", err.Error())
		return nil
	}
	records, err := c.readRecordsFromCsv()
	if err != nil {
		fmt.Println("Could not retrieve highscore: ", err.Error())
		return nil
	}
	if len(records) == 0 {
		return nil
	}
	return &records[0]
}

// Returns DESC sorted list of tracked score records
func (c *CsvScoreBoard) readRecordsFromCsv() ([]ScoreRecord, error) {
	f, err := os.Open(c.path)
	if err != nil {
		return []ScoreRecord{}, err
	}
	defer f.Close()

	// Parse CSV score data
	csvReader := csv.NewReader(f)
	records := []ScoreRecord{}
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return []ScoreRecord{}, err
		}
		if len(rec) != 2 {
			return []ScoreRecord{}, fmt.Errorf("Invalid number of cols received")
		}
		timestamp, err := strconv.ParseInt(rec[0], 10, 64)
		if err != nil {
			return []ScoreRecord{}, err
		}
		score, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			return []ScoreRecord{}, err
		}
		records = append(records, ScoreRecord{timestamp, ScoreValue(score)})
	}

	// Sort by score DESC
	slices.SortFunc(records, func(a, b ScoreRecord) int {
		return a.compareto(&b)
	})
	return records, nil
}

func (c *CsvScoreBoard) Print() error {
	scores, err := c.readRecordsFromCsv()
	if err != nil {
		return err
	}
	fmt.Println("--- CSV Scoreboard ---")
	for idx, score := range scores {
		fmt.Printf("[%d]: %s\n", idx, score.String())
	}
	return nil

}
