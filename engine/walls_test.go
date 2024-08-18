package engine_test

import (
	"fmt"
	"testing"

	"github.com/lucb31/game-engine-go/bin/assets"
	"github.com/lucb31/game-engine-go/engine"
)

func TestCsvParse(t *testing.T) {
	csv := assets.MapTDBaseWallsCSV
	mapData, err := engine.ReadCsvFromBinary(csv)
	if err != nil {
		t.Fatalf("Could not parse csv")
	}

	expectedHorizontalSegments := 6
	hSegments := engine.CalcHorizontalWallSegments(mapData)
	fmt.Println("Horizontal segments: ", hSegments)
	if len(hSegments) != expectedHorizontalSegments {
		t.Fatalf("Invalid number of horizontal segments. Expected %d, but recevied %d", expectedHorizontalSegments, len(hSegments))
	}

	expectedVerticalSegments := 6
	vSegments := engine.CalcVerticalWallSegments(mapData)
	fmt.Println("Vertical segments", vSegments)
	if len(vSegments) != expectedVerticalSegments {
		t.Fatalf("Invalid number of veritcal segments. Expected %d, but recevied %d", expectedHorizontalSegments, len(hSegments))
	}
}
