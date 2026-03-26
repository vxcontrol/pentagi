package helpers

import (
	"testing"

	"pentagi/pkg/csum"
)

// TestConfigBoundaries verifies that boundaries are calculated correctly based on configuration
func TestConfigBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		config csum.SummarizerConfig
		verify func(t *testing.T, boundaries ConfigBoundaries)
	}{
		{
			name: "Default configuration boundaries",
			config: csum.SummarizerConfig{
				PreserveLast:   true,
				UseQA:          true,
				SummHumanInQA:  false,
				LastSecBytes:   50 * 1024,
				MaxBPBytes:     16 * 1024,
				MaxQABytes:     100 * 1024,
				MaxQASections:  10,
				KeepQASections: 2,
			},
			verify: func(t *testing.T, b ConfigBoundaries) {
				// Check that boundaries respect configuration limits
				if b.MaxSectionBytes > 50*1024 {
					t.Errorf("MaxSectionBytes (%d) should not exceed LastSecBytes (50KB)", b.MaxSectionBytes)
				}
				if b.MaxBodyPairBytes > 16*1024 {
					t.Errorf("MaxBodyPairBytes (%d) should not exceed MaxBPBytes (16KB)", b.MaxBodyPairBytes)
				}
				if b.MaxQABytes > 100*1024 {
					t.Errorf("MaxQABytes (%d) should not exceed config MaxQABytes (100KB)", b.MaxQABytes)
				}

				// Check minimum bounds
				if b.MinSectionBytes < MinSectionBytes {
					t.Errorf("MinSectionBytes (%d) below absolute minimum (%d)", b.MinSectionBytes, MinSectionBytes)
				}
				if b.MinBodyPairBytes < MinBodyPairBytes {
					t.Errorf("MinBodyPairBytes (%d) below absolute minimum (%d)", b.MinBodyPairBytes, MinBodyPairBytes)
				}

				// Check logical consistency
				if b.MaxSectionBytes < b.MinSectionBytes {
					t.Errorf("MaxSectionBytes (%d) < MinSectionBytes (%d)", b.MaxSectionBytes, b.MinSectionBytes)
				}
				if b.MaxBodyPairBytes < b.MinBodyPairBytes {
					t.Errorf("MaxBodyPairBytes (%d) < MinBodyPairBytes (%d)", b.MaxBodyPairBytes, b.MinBodyPairBytes)
				}
			},
		},
		{
			name: "Extreme low values boundaries",
			config: csum.SummarizerConfig{
				PreserveLast:   true,
				UseQA:          true,
				SummHumanInQA:  false,
				LastSecBytes:   10 * 1024, // Very low
				MaxBPBytes:     4 * 1024,  // Very low
				MaxQABytes:     20 * 1024, // Very low
				MaxQASections:  2,         // Very low
				KeepQASections: 1,
			},
			verify: func(t *testing.T, b ConfigBoundaries) {
				// Even with low config values, boundaries should not go below reasonable minimums
				if b.MinSectionBytes < ReasonableMinSectionBytes {
					t.Errorf("MinSectionBytes (%d) should be at least reasonable minimum (%d)",
						b.MinSectionBytes, ReasonableMinSectionBytes)
				}
				if b.MinBodyPairBytes < ReasonableMinBodyPairBytes {
					t.Errorf("MinBodyPairBytes (%d) should be at least reasonable minimum (%d)",
						b.MinBodyPairBytes, ReasonableMinBodyPairBytes)
				}
			},
		},
		{
			name: "Extreme high values boundaries",
			config: csum.SummarizerConfig{
				PreserveLast:   true,
				UseQA:          true,
				SummHumanInQA:  false,
				LastSecBytes:   200 * 1024,  // Very high
				MaxBPBytes:     64 * 1024,   // Very high
				MaxQABytes:     1024 * 1024, // Very high
				MaxQASections:  50,          // Very high
				KeepQASections: 20,
			},
			verify: func(t *testing.T, b ConfigBoundaries) {
				// High config values should be capped at reasonable maximums
				if b.MaxSectionBytes > ReasonableMaxSectionBytes {
					t.Errorf("MaxSectionBytes (%d) should be capped at reasonable maximum (%d)",
						b.MaxSectionBytes, ReasonableMaxSectionBytes)
				}
				if b.MaxBodyPairBytes > ReasonableMaxBodyPairBytes {
					t.Errorf("MaxBodyPairBytes (%d) should be capped at reasonable maximum (%d)",
						b.MaxBodyPairBytes, ReasonableMaxBodyPairBytes)
				}
				if b.MaxQABytes > ReasonableMaxQABytes {
					t.Errorf("MaxQABytes (%d) should be capped at reasonable maximum (%d)",
						b.MaxQABytes, ReasonableMaxQABytes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boundaries := NewConfigBoundaries(tt.config)

			t.Logf("Config: LastSec=%dKB, MaxBP=%dKB, MaxQA=%dKB, MaxQASections=%d, KeepQA=%d",
				tt.config.LastSecBytes/1024, tt.config.MaxBPBytes/1024, tt.config.MaxQABytes/1024,
				tt.config.MaxQASections, tt.config.KeepQASections)
			t.Logf("Boundaries: MinSec=%dKB, MaxSec=%dKB, MinBP=%dKB, MaxBP=%dKB, MinQA=%dKB, MaxQA=%dKB",
				boundaries.MinSectionBytes/1024, boundaries.MaxSectionBytes/1024,
				boundaries.MinBodyPairBytes/1024, boundaries.MaxBodyPairBytes/1024,
				boundaries.MinQABytes/1024, boundaries.MaxQABytes/1024)

			tt.verify(t, boundaries)
		})
	}
}

// TestMonotonicBehavior tests that increasing each parameter never decreases context estimates
func TestMonotonicBehavior(t *testing.T) {
	baseConfig := csum.SummarizerConfig{
		PreserveLast:   true,
		UseQA:          true,
		SummHumanInQA:  false,
		LastSecBytes:   50 * 1024,
		MaxBPBytes:     16 * 1024,
		MaxQABytes:     100 * 1024,
		MaxQASections:  10,
		KeepQASections: 2,
	}

	// Test KeepQASections monotonicity (most important parameter)
	t.Run("KeepQASections", func(t *testing.T) {
		var prevEstimate *ContextEstimate
		for _, keepSections := range []int{1, 2, 3, 5, 7, 10} {
			config := baseConfig
			config.KeepQASections = keepSections
			config.MaxQASections = max(config.MaxQASections, keepSections) // Ensure consistency

			estimate := CalculateContextEstimate(config)
			t.Logf("KeepQASections=%d: Min=%d, Max=%d tokens",
				keepSections, estimate.MinTokens, estimate.MaxTokens)

			if prevEstimate != nil {
				if estimate.MinTokens < prevEstimate.MinTokens {
					t.Errorf("Non-monotonic MinTokens: %d < %d for KeepQASections %d",
						estimate.MinTokens, prevEstimate.MinTokens, keepSections)
				}
				if estimate.MaxTokens < prevEstimate.MaxTokens {
					t.Errorf("Non-monotonic MaxTokens: %d < %d for KeepQASections %d",
						estimate.MaxTokens, prevEstimate.MaxTokens, keepSections)
				}
			}
			prevEstimate = &estimate
		}
	})

	// Test LastSecBytes monotonicity
	t.Run("LastSecBytes", func(t *testing.T) {
		var prevEstimate *ContextEstimate
		for _, lastSecBytes := range []int{20 * 1024, 30 * 1024, 50 * 1024, 70 * 1024, 100 * 1024} {
			config := baseConfig
			config.LastSecBytes = lastSecBytes

			estimate := CalculateContextEstimate(config)
			t.Logf("LastSecBytes=%dKB: Min=%d, Max=%d tokens",
				lastSecBytes/1024, estimate.MinTokens, estimate.MaxTokens)

			if prevEstimate != nil {
				if estimate.MinTokens < prevEstimate.MinTokens {
					t.Errorf("Non-monotonic MinTokens: %d < %d for LastSecBytes %dKB",
						estimate.MinTokens, prevEstimate.MinTokens, lastSecBytes/1024)
				}
				if estimate.MaxTokens < prevEstimate.MaxTokens {
					t.Errorf("Non-monotonic MaxTokens: %d < %d for LastSecBytes %dKB",
						estimate.MaxTokens, prevEstimate.MaxTokens, lastSecBytes/1024)
				}
			}
			prevEstimate = &estimate
		}
	})

	// Test MaxQABytes monotonicity
	t.Run("MaxQABytes", func(t *testing.T) {
		var prevEstimate *ContextEstimate
		for _, maxQABytes := range []int{50 * 1024, 75 * 1024, 100 * 1024, 150 * 1024, 200 * 1024} {
			config := baseConfig
			config.MaxQABytes = maxQABytes

			estimate := CalculateContextEstimate(config)
			t.Logf("MaxQABytes=%dKB: Min=%d, Max=%d tokens",
				maxQABytes/1024, estimate.MinTokens, estimate.MaxTokens)

			if prevEstimate != nil {
				if estimate.MinTokens < prevEstimate.MinTokens {
					t.Errorf("Non-monotonic MinTokens: %d < %d for MaxQABytes %dKB",
						estimate.MinTokens, prevEstimate.MinTokens, maxQABytes/1024)
				}
				if estimate.MaxTokens < prevEstimate.MaxTokens {
					t.Errorf("Non-monotonic MaxTokens: %d < %d for MaxQABytes %dKB",
						estimate.MaxTokens, prevEstimate.MaxTokens, maxQABytes/1024)
				}
			}
			prevEstimate = &estimate
		}
	})

	// Test MaxQASections monotonicity
	t.Run("MaxQASections", func(t *testing.T) {
		var prevEstimate *ContextEstimate
		for _, maxQASections := range []int{3, 5, 8, 10, 15} {
			config := baseConfig
			config.MaxQASections = maxQASections

			estimate := CalculateContextEstimate(config)
			t.Logf("MaxQASections=%d: Min=%d, Max=%d tokens",
				maxQASections, estimate.MinTokens, estimate.MaxTokens)

			if prevEstimate != nil {
				if estimate.MinTokens < prevEstimate.MinTokens {
					t.Errorf("Non-monotonic MinTokens: %d < %d for MaxQASections %d",
						estimate.MinTokens, prevEstimate.MinTokens, maxQASections)
				}
				if estimate.MaxTokens < prevEstimate.MaxTokens {
					t.Errorf("Non-monotonic MaxTokens: %d < %d for MaxQASections %d",
						estimate.MaxTokens, prevEstimate.MaxTokens, maxQASections)
				}
			}
			prevEstimate = &estimate
		}
	})

	// Test MaxBPBytes monotonicity
	t.Run("MaxBPBytes", func(t *testing.T) {
		var prevEstimate *ContextEstimate
		for _, maxBPBytes := range []int{8 * 1024, 12 * 1024, 16 * 1024, 24 * 1024, 32 * 1024} {
			config := baseConfig
			config.MaxBPBytes = maxBPBytes

			estimate := CalculateContextEstimate(config)
			t.Logf("MaxBPBytes=%dKB: Min=%d, Max=%d tokens",
				maxBPBytes/1024, estimate.MinTokens, estimate.MaxTokens)

			if prevEstimate != nil {
				if estimate.MinTokens < prevEstimate.MinTokens {
					t.Errorf("Non-monotonic MinTokens: %d < %d for MaxBPBytes %dKB",
						estimate.MinTokens, prevEstimate.MinTokens, maxBPBytes/1024)
				}
				if estimate.MaxTokens < prevEstimate.MaxTokens {
					t.Errorf("Non-monotonic MaxTokens: %d < %d for MaxBPBytes %dKB",
						estimate.MaxTokens, prevEstimate.MaxTokens, maxBPBytes/1024)
				}
			}
			prevEstimate = &estimate
		}
	})
}

// TestBooleanParametersLogic tests correct behavior of boolean parameters
func TestBooleanParametersLogic(t *testing.T) {
	baseConfig := csum.SummarizerConfig{
		LastSecBytes:   50 * 1024,
		MaxBPBytes:     16 * 1024,
		MaxQABytes:     100 * 1024,
		MaxQASections:  10,
		KeepQASections: 3,
	}

	// Test PreserveLast parameter (CRITICAL TEST)
	t.Run("PreserveLast", func(t *testing.T) {
		configFalse := baseConfig
		configFalse.PreserveLast = false
		configFalse.UseQA = true
		configFalse.SummHumanInQA = false

		configTrue := baseConfig
		configTrue.PreserveLast = true
		configTrue.UseQA = true
		configTrue.SummHumanInQA = false

		estimateFalse := CalculateContextEstimate(configFalse)
		estimateTrue := CalculateContextEstimate(configTrue)

		t.Logf("PreserveLast=false: Min=%d, Max=%d tokens", estimateFalse.MinTokens, estimateFalse.MaxTokens)
		t.Logf("PreserveLast=true: Min=%d, Max=%d tokens", estimateTrue.MinTokens, estimateTrue.MaxTokens)

		// CRITICAL: PreserveLast=true should result in SMALLER context (more summarization)
		// PreserveLast=false should result in LARGER context (less summarization)
		if estimateTrue.MaxTokens >= estimateFalse.MaxTokens {
			t.Errorf("PreserveLast=true should produce SMALLER MaxTokens than false. Got true=%d, false=%d",
				estimateTrue.MaxTokens, estimateFalse.MaxTokens)
		}

		if estimateTrue.MinTokens > estimateFalse.MinTokens {
			t.Errorf("PreserveLast=true should produce SMALLER or equal MinTokens than false. Got true=%d, false=%d",
				estimateTrue.MinTokens, estimateFalse.MinTokens)
		}
	})

	// Test UseQA parameter
	t.Run("UseQA", func(t *testing.T) {
		configFalse := baseConfig
		configFalse.PreserveLast = true
		configFalse.UseQA = false
		configFalse.SummHumanInQA = false

		configTrue := baseConfig
		configTrue.PreserveLast = true
		configTrue.UseQA = true
		configTrue.SummHumanInQA = false

		estimateFalse := CalculateContextEstimate(configFalse)
		estimateTrue := CalculateContextEstimate(configTrue)

		t.Logf("UseQA=false: Min=%d, Max=%d tokens", estimateFalse.MinTokens, estimateFalse.MaxTokens)
		t.Logf("UseQA=true: Min=%d, Max=%d tokens", estimateTrue.MinTokens, estimateTrue.MaxTokens)

		// UseQA should affect the results (direction depends on scenario, but should be different)
		if estimateFalse.MinTokens == estimateTrue.MinTokens && estimateFalse.MaxTokens == estimateTrue.MaxTokens {
			t.Errorf("UseQA parameter should affect the estimates")
		}
	})

	// Test SummHumanInQA parameter
	t.Run("SummHumanInQA", func(t *testing.T) {
		configFalse := baseConfig
		configFalse.PreserveLast = true
		configFalse.UseQA = true
		configFalse.SummHumanInQA = false

		configTrue := baseConfig
		configTrue.PreserveLast = true
		configTrue.UseQA = true
		configTrue.SummHumanInQA = true

		estimateFalse := CalculateContextEstimate(configFalse)
		estimateTrue := CalculateContextEstimate(configTrue)

		t.Logf("SummHumanInQA=false: Min=%d, Max=%d tokens", estimateFalse.MinTokens, estimateFalse.MaxTokens)
		t.Logf("SummHumanInQA=true: Min=%d, Max=%d tokens", estimateTrue.MinTokens, estimateTrue.MaxTokens)

		// SummHumanInQA=true should result in smaller context (more summarization)
		if estimateTrue.MaxTokens > estimateFalse.MaxTokens {
			t.Errorf("SummHumanInQA=true should produce smaller or equal MaxTokens than false. Got true=%d, false=%d",
				estimateTrue.MaxTokens, estimateFalse.MaxTokens)
		}
	})
}

// TestBoundariesUsage verifies that calculation functions actually use the boundaries
func TestBoundariesUsage(t *testing.T) {
	// This test ensures that boundaries are actually used in calculations
	config1 := csum.SummarizerConfig{
		PreserveLast:   true,
		UseQA:          true,
		SummHumanInQA:  false,
		LastSecBytes:   30 * 1024, // Low value
		MaxBPBytes:     8 * 1024,  // Low value
		MaxQABytes:     50 * 1024, // Low value
		MaxQASections:  5,
		KeepQASections: 2,
	}

	config2 := csum.SummarizerConfig{
		PreserveLast:   true,
		UseQA:          true,
		SummHumanInQA:  false,
		LastSecBytes:   80 * 1024,  // High value
		MaxBPBytes:     24 * 1024,  // High value
		MaxQABytes:     200 * 1024, // High value
		MaxQASections:  5,          // Same as config1
		KeepQASections: 2,          // Same as config1
	}

	boundaries1 := NewConfigBoundaries(config1)
	boundaries2 := NewConfigBoundaries(config2)

	// Boundaries should be different
	if boundaries1.MaxSectionBytes == boundaries2.MaxSectionBytes {
		t.Errorf("Boundaries should differ based on configuration")
	}

	estimate1 := CalculateContextEstimate(config1)
	estimate2 := CalculateContextEstimate(config2)

	t.Logf("Config1 boundaries: MaxSec=%dKB, MaxBP=%dKB, MaxQA=%dKB",
		boundaries1.MaxSectionBytes/1024, boundaries1.MaxBodyPairBytes/1024, boundaries1.MaxQABytes/1024)
	t.Logf("Config2 boundaries: MaxSec=%dKB, MaxBP=%dKB, MaxQA=%dKB",
		boundaries2.MaxSectionBytes/1024, boundaries2.MaxBodyPairBytes/1024, boundaries2.MaxQABytes/1024)

	t.Logf("Config1 estimate: Min=%d, Max=%d tokens", estimate1.MinTokens, estimate1.MaxTokens)
	t.Logf("Config2 estimate: Min=%d, Max=%d tokens", estimate2.MinTokens, estimate2.MaxTokens)

	// Config2 should have larger estimates since it has larger limits
	if estimate2.MaxTokens <= estimate1.MaxTokens {
		t.Errorf("Config with larger limits should produce larger estimates. Got config1=%d, config2=%d",
			estimate1.MaxTokens, estimate2.MaxTokens)
	}
}

// TestCalculateContextEstimate verifies the main function works correctly
func TestCalculateContextEstimate(t *testing.T) {
	testCases := []struct {
		name   string
		config csum.SummarizerConfig
		verify func(t *testing.T, estimate ContextEstimate)
	}{
		{
			name: "Minimal configuration",
			config: csum.SummarizerConfig{
				PreserveLast:   false,
				UseQA:          false,
				SummHumanInQA:  false,
				LastSecBytes:   20 * 1024,
				MaxBPBytes:     8 * 1024,
				MaxQABytes:     30 * 1024,
				MaxQASections:  3,
				KeepQASections: 1,
			},
			verify: func(t *testing.T, estimate ContextEstimate) {
				if estimate.MinTokens <= 0 || estimate.MaxTokens <= estimate.MinTokens {
					t.Errorf("Invalid estimates: Min=%d, Max=%d", estimate.MinTokens, estimate.MaxTokens)
				}
				// Should be relatively small
				if estimate.MaxTokens > 20000 {
					t.Errorf("Minimal config should produce modest estimates, got %d tokens", estimate.MaxTokens)
				}
			},
		},
		{
			name: "Maximal configuration",
			config: csum.SummarizerConfig{
				PreserveLast:   true,
				UseQA:          true,
				SummHumanInQA:  true,
				LastSecBytes:   100 * 1024,
				MaxBPBytes:     32 * 1024,
				MaxQABytes:     400 * 1024,
				MaxQASections:  15,
				KeepQASections: 8,
			},
			verify: func(t *testing.T, estimate ContextEstimate) {
				if estimate.MinTokens <= 0 || estimate.MaxTokens <= estimate.MinTokens {
					t.Errorf("Invalid estimates: Min=%d, Max=%d", estimate.MinTokens, estimate.MaxTokens)
				}
				// Should be larger than minimal config
				if estimate.MaxTokens < 30000 {
					t.Errorf("Maximal config should produce substantial estimates, got %d tokens", estimate.MaxTokens)
				}
			},
		},
		{
			name: "Default-like configuration",
			config: csum.SummarizerConfig{
				PreserveLast:   true,
				UseQA:          true,
				SummHumanInQA:  false,
				LastSecBytes:   50 * 1024,
				MaxBPBytes:     16 * 1024,
				MaxQABytes:     64 * 1024,
				MaxQASections:  10,
				KeepQASections: 1,
			},
			verify: func(t *testing.T, estimate ContextEstimate) {
				// Check token to byte ratio
				expectedMinBytes := estimate.MinTokens * TokenToByteRatio
				expectedMaxBytes := estimate.MaxTokens * TokenToByteRatio

				// Allow small rounding errors (due to divisions in calculation)
				if abs(estimate.MinBytes-expectedMinBytes) > 5 {
					t.Errorf("MinBytes calculation error: expected %d, got %d", expectedMinBytes, estimate.MinBytes)
				}
				if abs(estimate.MaxBytes-expectedMaxBytes) > 5 {
					t.Errorf("MaxBytes calculation error: expected %d, got %d", expectedMaxBytes, estimate.MaxBytes)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estimate := CalculateContextEstimate(tc.config)

			t.Logf("%s: Min=%d tokens (%d bytes), Max=%d tokens (%d bytes)",
				tc.name, estimate.MinTokens, estimate.MinBytes, estimate.MaxTokens, estimate.MaxBytes)

			tc.verify(t, estimate)
		})
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
