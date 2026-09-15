package models

import "testing"

func TestParallelSearchLogEngine(t *testing.T) {
	if err := SearchEngineTypeParallel.Valid(); err != nil {
		t.Fatalf("Parallel search attribution rejected: %v", err)
	}
	if err := SearchEngineType("unknown").Valid(); err == nil {
		t.Fatal("unknown engine accepted")
	}
}
