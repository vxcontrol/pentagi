package models

import "testing"

func TestSearchEngineTypeXquikValid(t *testing.T) {
	if err := SearchEngineTypeXquik.Valid(); err != nil {
		t.Fatalf("SearchEngineTypeXquik.Valid() error: %v", err)
	}
}
