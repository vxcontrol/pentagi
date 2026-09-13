package tools

import (
	"strings"
	"testing"
)

func TestTruncateForEmbedding_UnderLimit(t *testing.T) {
	t.Parallel()

	text := "short text"
	result := truncateForEmbedding(text, 100)

	if result != text {
		t.Errorf("expected text unchanged when under limit, got %q", result)
	}
}

func TestTruncateForEmbedding_ExactLimit(t *testing.T) {
	t.Parallel()

	text := "exactly10c"
	if len(text) != 10 {
		t.Fatalf("test fixture length changed: got %d", len(text))
	}

	result := truncateForEmbedding(text, 10)

	if result != text {
		t.Errorf("expected text unchanged when exactly at limit, got %q", result)
	}
}

func TestTruncateForEmbedding_OverLimit(t *testing.T) {
	t.Parallel()

	text := "this text is definitely longer than the limit"
	result := truncateForEmbedding(text, 10)

	if len(result) != 10 {
		t.Errorf("expected truncated length 10, got %d (%q)", len(result), result)
	}
	if result != text[:10] {
		t.Errorf("expected prefix of original text, got %q", result)
	}
}

func TestTruncateForEmbedding_ZeroOrNegativeLimit(t *testing.T) {
	t.Parallel()

	text := "some text"

	if result := truncateForEmbedding(text, 0); result != text {
		t.Errorf("maxBytes=0 should return text unchanged, got %q", result)
	}
	if result := truncateForEmbedding(text, -1); result != text {
		t.Errorf("negative maxBytes should return text unchanged, got %q", result)
	}
}

func TestTruncateForEmbedding_EmptyText(t *testing.T) {
	t.Parallel()

	if result := truncateForEmbedding("", 10); result != "" {
		t.Errorf("expected empty string unchanged, got %q", result)
	}
}

func TestFormatVectorFromFloat32s_EmptySlice(t *testing.T) {
	t.Parallel()

	result := formatVectorFromFloat32s([]float32{})

	if result != "[]" {
		t.Errorf("expected empty vector literal '[]', got %q", result)
	}
}

func TestFormatVectorFromFloat32s_SingleValue(t *testing.T) {
	t.Parallel()

	result := formatVectorFromFloat32s([]float32{1.5})

	if result != "[1.5]" {
		t.Errorf("expected '[1.5]', got %q", result)
	}
}

func TestFormatVectorFromFloat32s_MultipleValues(t *testing.T) {
	t.Parallel()

	result := formatVectorFromFloat32s([]float32{1, 2.25, -3.5})

	if result != "[1,2.25,-3.5]" {
		t.Errorf("expected '[1,2.25,-3.5]', got %q", result)
	}
}

func TestFormatVectorFromFloat32s_IsCommaSeparatedAndBracketed(t *testing.T) {
	t.Parallel()

	result := formatVectorFromFloat32s([]float32{0.1, 0.2, 0.3})

	if !strings.HasPrefix(result, "[") || !strings.HasSuffix(result, "]") {
		t.Errorf("expected result to be wrapped in brackets, got %q", result)
	}
	if strings.Count(result, ",") != 2 {
		t.Errorf("expected 2 commas for 3 values, got %q", result)
	}
}
