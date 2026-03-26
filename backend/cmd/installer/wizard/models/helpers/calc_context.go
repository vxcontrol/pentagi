package helpers

import (
	"pentagi/pkg/csum"
)

// ContextEstimate represents the estimated context size range
type ContextEstimate struct {
	MinTokens int // Minimum estimated tokens (optimal summarization)
	MaxTokens int // Maximum estimated tokens (approaching limits)
	MinBytes  int // Minimum estimated bytes
	MaxBytes  int // Maximum estimated bytes
}

// Global parameter boundaries based on data nature and algorithm constraints
var (
	// Absolute minimums based on data nature
	MinBodyPairBytes      = 512      // Minimum possible body pair size
	MinSectionBytes       = 3 * 1024 // Minimum section: header + 1 body pair
	MinSystemMessageBytes = 1 * 1024 // Minimum system message
	MinHumanMessageBytes  = 512      // Minimum human message
	MinQASections         = 1        // At least one QA section
	MinKeepQASections     = 1        // At least one section to keep

	// Typical sizes for realistic estimation
	TypicalSystemMessageBytes = 4 * 1024 // ~1k tokens for system message
	TypicalHumanMessageBytes  = 2 * 1024 // ~512 tokens for human message
	TypicalBodyPairBytes      = 8 * 1024 // ~2k tokens typical body pair
	SummarizedBodyPairBytes   = 6 * 1024 // ~1.5k tokens after summarization
	QASummaryHeaderBytes      = 4 * 1024 // ~1k tokens for QA summary header
	QASummaryBodyPairBytes    = 8 * 1024 // ~2k tokens for QA summary body pair

	// Reasonable ranges where parameters still have meaningful impact
	ReasonableMinBodyPairBytes = 8 * 1024   // 8KB
	ReasonableMaxBodyPairBytes = 32 * 1024  // 32KB
	ReasonableMinSectionBytes  = 15 * 1024  // 15KB
	ReasonableMaxSectionBytes  = 100 * 1024 // 100KB
	ReasonableMinQABytes       = 30 * 1024  // 30KB
	ReasonableMaxQABytes       = 500 * 1024 // 500KB
	ReasonableMinQASections    = 2          // 2 sections
	ReasonableMaxQASections    = 15         // 15 sections

	// Token to byte conversion ratio
	TokenToByteRatio = 4
)

// ConfigBoundaries represents effective boundaries for a specific configuration
type ConfigBoundaries struct {
	// Effective ranges for parameters based on configuration
	MinBodyPairBytes int
	MaxBodyPairBytes int
	MinSectionBytes  int
	MaxSectionBytes  int
	MinQABytes       int
	MaxQABytes       int
	MinQASections    int
	MaxQASections    int
	MinKeepSections  int
	MaxKeepSections  int

	// Derived boundaries
	MinSectionsToProcess int // Minimum sections that would be processed
	MaxSectionsToProcess int // Maximum sections before QA triggers
}

// NewConfigBoundaries creates boundaries adjusted for specific configuration
func NewConfigBoundaries(config csum.SummarizerConfig) ConfigBoundaries {
	boundaries := ConfigBoundaries{}

	// Body pair boundaries
	boundaries.MinBodyPairBytes = max(MinBodyPairBytes, ReasonableMinBodyPairBytes)
	if config.MaxBPBytes > 0 {
		boundaries.MaxBodyPairBytes = min(config.MaxBPBytes, ReasonableMaxBodyPairBytes)
	} else {
		boundaries.MaxBodyPairBytes = ReasonableMaxBodyPairBytes
	}
	boundaries.MaxBodyPairBytes = max(boundaries.MaxBodyPairBytes, boundaries.MinBodyPairBytes)

	// Section boundaries
	boundaries.MinSectionBytes = max(MinSectionBytes, ReasonableMinSectionBytes)
	if config.LastSecBytes > 0 {
		boundaries.MaxSectionBytes = min(config.LastSecBytes, ReasonableMaxSectionBytes)
	} else {
		boundaries.MaxSectionBytes = ReasonableMaxSectionBytes
	}
	boundaries.MaxSectionBytes = max(boundaries.MaxSectionBytes, boundaries.MinSectionBytes)

	// QA bytes boundaries
	boundaries.MinQABytes = max(boundaries.MinSectionBytes*ReasonableMinQASections, ReasonableMinQABytes)
	if config.MaxQABytes > 0 {
		boundaries.MaxQABytes = min(config.MaxQABytes, ReasonableMaxQABytes)
	} else {
		boundaries.MaxQABytes = ReasonableMaxQABytes
	}
	boundaries.MaxQABytes = max(boundaries.MaxQABytes, boundaries.MinQABytes)

	// QA sections boundaries
	boundaries.MinQASections = max(MinQASections, ReasonableMinQASections)
	if config.MaxQASections > 0 {
		boundaries.MaxQASections = min(config.MaxQASections, ReasonableMaxQASections)
	} else {
		boundaries.MaxQASections = ReasonableMaxQASections
	}
	boundaries.MaxQASections = max(boundaries.MaxQASections, boundaries.MinQASections)

	// Keep sections boundaries
	boundaries.MinKeepSections = max(MinKeepQASections, config.KeepQASections)
	boundaries.MaxKeepSections = min(boundaries.MaxQASections, config.KeepQASections)
	boundaries.MaxKeepSections = max(boundaries.MaxKeepSections, boundaries.MinKeepSections)

	// Derived boundaries for sections processing
	boundaries.MinSectionsToProcess = boundaries.MinKeepSections + 1

	// Calculate when QA summarization would trigger
	minSectionSize := boundaries.MinSectionBytes
	maxSectionsBeforeQA := boundaries.MaxQABytes / minSectionSize
	boundaries.MaxSectionsToProcess = min(maxSectionsBeforeQA, boundaries.MaxQASections)
	boundaries.MaxSectionsToProcess = max(boundaries.MaxSectionsToProcess, boundaries.MinSectionsToProcess)

	return boundaries
}

// CalculateContextEstimate calculates the estimated context size based on summarizer configuration
func CalculateContextEstimate(config csum.SummarizerConfig) ContextEstimate {
	// Create boundaries for this configuration
	boundaries := NewConfigBoundaries(config)

	// Calculate minimum context (optimal summarization scenario)
	minBytes := calculateMinimumContext(config, boundaries)

	// Calculate maximum context (approaching limits scenario)
	maxBytes := calculateMaximumContext(config, boundaries)

	// Convert bytes to tokens
	minTokens := minBytes / TokenToByteRatio
	maxTokens := maxBytes / TokenToByteRatio

	return ContextEstimate{
		MinTokens: minTokens,
		MaxTokens: maxTokens,
		MinBytes:  minBytes,
		MaxBytes:  maxBytes,
	}
}

// calculateMinimumContext estimates minimum context when summarization works optimally
func calculateMinimumContext(config csum.SummarizerConfig, boundaries ConfigBoundaries) int {
	totalBytes := 0

	// Base overhead: system message
	totalBytes += TypicalSystemMessageBytes

	// Base sections: use boundaries for minimum sections count
	baseSections := max(boundaries.MinKeepSections, 1)

	// Use boundaries for minimum section size calculation
	minSectionSize := TypicalHumanMessageBytes + SummarizedBodyPairBytes

	// For each base section, calculate minimal content
	for i := 0; i < baseSections; i++ {
		// Section header
		totalBytes += TypicalHumanMessageBytes

		// Section body - use minimal body pair sizes from boundaries
		var sectionBodySize int
		if config.PreserveLast {
			// PreserveLast=true means sections are summarized to fit LastSecBytes
			// This REDUCES total size (more compression)
			sectionBodySize = min(boundaries.MinSectionBytes/3, SummarizedBodyPairBytes*2)
		} else {
			// PreserveLast=false means sections remain as-is without last section management
			// This INCREASES total size (less compression)
			sectionBodySize = min(boundaries.MinSectionBytes, TypicalBodyPairBytes*2)
		}
		totalBytes += sectionBodySize
	}

	// Add minimal impact from configuration parameters using boundaries

	// MaxBPBytes influence through boundaries
	if boundaries.MaxBodyPairBytes > boundaries.MinBodyPairBytes {
		// Add small fraction of the difference for minimal scenario
		overhead := (boundaries.MaxBodyPairBytes - boundaries.MinBodyPairBytes) / 10
		totalBytes += overhead
	}

	// MaxQABytes influence through boundaries
	if boundaries.MaxQABytes > boundaries.MinQABytes {
		// Add small fraction for potential QA content
		qaOverhead := (boundaries.MaxQABytes - boundaries.MinQABytes) / 20
		totalBytes += qaOverhead
	}

	// MaxQASections influence through boundaries
	if boundaries.MaxQASections > boundaries.MinQASections {
		// Each additional section beyond minimum adds fractional content
		additionalSections := boundaries.MaxQASections - boundaries.MinQASections
		totalBytes += min(additionalSections, 3) * (minSectionSize / 4)
	}

	// Boolean parameter influences for optimization
	if config.UseQA {
		// QA summarization generally reduces context through better summarization
		totalBytes = totalBytes * 9 / 10 // 10% reduction

		if config.SummHumanInQA {
			// Summarizing human messages saves additional space
			totalBytes = totalBytes * 95 / 100 // Additional 5% reduction
		}
	}

	// Note: PreserveLast effect is already built into section calculation above

	return totalBytes
}

// calculateMaximumContext estimates maximum context when approaching configuration limits
func calculateMaximumContext(config csum.SummarizerConfig, boundaries ConfigBoundaries) int {
	totalBytes := 0

	// Base overhead: system message
	totalBytes += TypicalSystemMessageBytes

	// Base sections: use boundaries for maximum sections that would be processed
	baseSections := max(boundaries.MaxKeepSections, boundaries.MinKeepSections)

	// Use boundaries for maximum section size calculation
	maxSectionSize := TypicalHumanMessageBytes + TypicalBodyPairBytes

	// For each base section, calculate maximum content
	for i := 0; i < baseSections; i++ {
		// Section header
		totalBytes += TypicalHumanMessageBytes

		// Section body - use maximum body pair sizes from boundaries
		var sectionBodySize int
		if config.PreserveLast {
			// PreserveLast=true means sections are summarized to fit within LastSecBytes
			// This keeps size SMALLER (more compression)
			sectionBodySize = min(boundaries.MaxSectionBytes, TypicalBodyPairBytes*3)
		} else {
			// PreserveLast=false means sections can grow larger without management
			// This allows size to be LARGER (less compression)
			sectionBodySize = min(boundaries.MaxSectionBytes*2, ReasonableMaxSectionBytes)
		}
		totalBytes += sectionBodySize
	}

	// Add maximum impact from configuration parameters using boundaries

	// MaxBPBytes influence through boundaries - full impact in maximum scenario
	if boundaries.MaxBodyPairBytes > boundaries.MinBodyPairBytes {
		// Add significant portion of the difference for maximum scenario
		overhead := (boundaries.MaxBodyPairBytes - boundaries.MinBodyPairBytes) / 2
		totalBytes += overhead
	}

	// MaxQABytes influence through boundaries - substantial impact
	if boundaries.MaxQABytes > boundaries.MinQABytes {
		// Add larger fraction for maximum QA content potential
		qaOverhead := (boundaries.MaxQABytes - boundaries.MinQABytes) / 8
		totalBytes += qaOverhead
	}

	// MaxQASections influence through boundaries - linear growth
	if boundaries.MaxQASections > boundaries.MinQASections {
		// Each additional section beyond minimum adds substantial content in max scenario
		additionalSections := boundaries.MaxQASections - boundaries.MinQASections
		totalBytes += min(additionalSections, 8) * (maxSectionSize / 2) // Half section size per additional
	}

	// Boolean parameter influences - mainly affecting complexity/growth
	if config.UseQA {
		// QA can enable more complex conversations in maximum scenario
		totalBytes = totalBytes * 105 / 100 // 5% increase for QA complexity

		if config.SummHumanInQA {
			// Summarizing human messages reduces maximum size
			totalBytes = totalBytes * 95 / 100 // 5% reduction
		}
	} else {
		// Without QA, conversations can be less organized and grow larger
		totalBytes = totalBytes * 110 / 100 // 10% increase
	}

	// KeepQASections direct influence: more kept sections = linearly more content
	// Use boundaries to ensure consistency
	keepSections := max(boundaries.MaxKeepSections, 1)
	if keepSections > 1 {
		// Each additional kept section adds substantial content in max scenario
		additionalKeptSections := keepSections - 1
		additionalContent := additionalKeptSections * maxSectionSize

		// Apply PreserveLast effect to additional content
		if config.PreserveLast {
			// PreserveLast reduces the size of additional content
			additionalContent = additionalContent * 8 / 10 // 20% reduction
		}

		totalBytes += additionalContent
	}

	// Add small buffer for message structure overhead
	totalBytes += 2 * 1024

	return totalBytes
}
