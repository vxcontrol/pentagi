package tools

import (
	"testing"

	"github.com/vxcontrol/langchaingo/schema"
)

func TestMergeAndDeduplicateDocs_EmptyInput(t *testing.T) {
	t.Parallel()

	result := MergeAndDeduplicateDocs([]schema.Document{}, 10)

	if len(result) != 0 {
		t.Errorf("MergeAndDeduplicateDocs with empty input should return empty slice, got %d items", len(result))
	}
}

func TestMergeAndDeduplicateDocs_NoDuplicates(t *testing.T) {
	t.Parallel()

	docs := []schema.Document{
		{PageContent: "content1", Score: 0.9, Metadata: map[string]any{"id": 1}},
		{PageContent: "content2", Score: 0.8, Metadata: map[string]any{"id": 2}},
		{PageContent: "content3", Score: 0.7, Metadata: map[string]any{"id": 3}},
	}

	result := MergeAndDeduplicateDocs(docs, 10)

	if len(result) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(result))
	}

	// Check sorting by score (descending)
	if result[0].Score != 0.9 {
		t.Errorf("First document should have highest score 0.9, got %f", result[0].Score)
	}
	if result[1].Score != 0.8 {
		t.Errorf("Second document should have score 0.8, got %f", result[1].Score)
	}
	if result[2].Score != 0.7 {
		t.Errorf("Third document should have score 0.7, got %f", result[2].Score)
	}
}

func TestMergeAndDeduplicateDocs_WithDuplicates(t *testing.T) {
	t.Parallel()

	docs := []schema.Document{
		{PageContent: "duplicate content", Score: 0.5, Metadata: map[string]any{"id": 1}},
		{PageContent: "unique content", Score: 0.8, Metadata: map[string]any{"id": 2}},
		{PageContent: "duplicate content", Score: 0.9, Metadata: map[string]any{"id": 3}}, // Higher score
		{PageContent: "another unique", Score: 0.7, Metadata: map[string]any{"id": 4}},
		{PageContent: "duplicate content", Score: 0.3, Metadata: map[string]any{"id": 5}}, // Lower score
	}

	result := MergeAndDeduplicateDocs(docs, 10)

	// Should have 3 unique documents
	if len(result) != 3 {
		t.Errorf("Expected 3 unique documents after deduplication, got %d", len(result))
	}

	// Find the "duplicate content" document
	var duplicateDoc *schema.Document
	for i := range result {
		if result[i].PageContent == "duplicate content" {
			duplicateDoc = &result[i]
			break
		}
	}

	if duplicateDoc == nil {
		t.Fatal("Duplicate content document not found in result")
	}

	// Should keep the one with highest score (0.9)
	if duplicateDoc.Score != 0.9 {
		t.Errorf("Duplicate document should have max score 0.9, got %f", duplicateDoc.Score)
	}

	// Should keep metadata from the document with highest score
	if duplicateDoc.Metadata["id"] != 3 {
		t.Errorf("Duplicate document should have metadata from doc with id=3, got %v", duplicateDoc.Metadata["id"])
	}
}

func TestMergeAndDeduplicateDocs_SortingByScore(t *testing.T) {
	t.Parallel()

	docs := []schema.Document{
		{PageContent: "content1", Score: 0.3, Metadata: map[string]any{}},
		{PageContent: "content2", Score: 0.9, Metadata: map[string]any{}},
		{PageContent: "content3", Score: 0.1, Metadata: map[string]any{}},
		{PageContent: "content4", Score: 0.7, Metadata: map[string]any{}},
		{PageContent: "content5", Score: 0.5, Metadata: map[string]any{}},
	}

	result := MergeAndDeduplicateDocs(docs, 10)

	// Check that results are sorted in descending order
	for i := 0; i < len(result)-1; i++ {
		if result[i].Score < result[i+1].Score {
			t.Errorf("Documents not sorted properly: result[%d].Score (%f) < result[%d].Score (%f)",
				i, result[i].Score, i+1, result[i+1].Score)
		}
	}

	// Verify exact order
	expectedScores := []float32{0.9, 0.7, 0.5, 0.3, 0.1}
	for i, expectedScore := range expectedScores {
		if result[i].Score != expectedScore {
			t.Errorf("result[%d].Score = %f, want %f", i, result[i].Score, expectedScore)
		}
	}
}

func TestMergeAndDeduplicateDocs_LimitEnforcement(t *testing.T) {
	t.Parallel()

	docs := []schema.Document{
		{PageContent: "content1", Score: 0.9, Metadata: map[string]any{}},
		{PageContent: "content2", Score: 0.8, Metadata: map[string]any{}},
		{PageContent: "content3", Score: 0.7, Metadata: map[string]any{}},
		{PageContent: "content4", Score: 0.6, Metadata: map[string]any{}},
		{PageContent: "content5", Score: 0.5, Metadata: map[string]any{}},
		{PageContent: "content6", Score: 0.4, Metadata: map[string]any{}},
		{PageContent: "content7", Score: 0.3, Metadata: map[string]any{}},
	}

	maxDocs := 3
	result := MergeAndDeduplicateDocs(docs, maxDocs)

	// Should return exactly maxDocs documents
	if len(result) != maxDocs {
		t.Errorf("Expected exactly %d documents, got %d", maxDocs, len(result))
	}

	// Should return documents with highest scores
	expectedScores := []float32{0.9, 0.8, 0.7}
	for i, expectedScore := range expectedScores {
		if result[i].Score != expectedScore {
			t.Errorf("result[%d].Score = %f, want %f (should select top scoring documents)",
				i, result[i].Score, expectedScore)
		}
	}
}

func TestMergeAndDeduplicateDocs_MetadataPreservation(t *testing.T) {
	t.Parallel()

	docs := []schema.Document{
		{
			PageContent: "same content",
			Score:       0.5,
			Metadata:    map[string]any{"source": "query1", "timestamp": "2023-01-01"},
		},
		{
			PageContent: "same content",
			Score:       0.9,
			Metadata:    map[string]any{"source": "query2", "timestamp": "2023-01-02"},
		},
		{
			PageContent: "same content",
			Score:       0.3,
			Metadata:    map[string]any{"source": "query3", "timestamp": "2023-01-03"},
		},
	}

	result := MergeAndDeduplicateDocs(docs, 10)

	if len(result) != 1 {
		t.Fatalf("Expected 1 deduplicated document, got %d", len(result))
	}

	// Should preserve metadata from document with highest score (0.9)
	if result[0].Metadata["source"] != "query2" {
		t.Errorf("Expected metadata from document with highest score, got source=%v", result[0].Metadata["source"])
	}
	if result[0].Metadata["timestamp"] != "2023-01-02" {
		t.Errorf("Expected timestamp from document with highest score, got %v", result[0].Metadata["timestamp"])
	}
}

func TestHashContent_Consistency(t *testing.T) {
	t.Parallel()

	content := "test content for hashing"

	hash1 := hashContent(content)
	hash2 := hashContent(content)

	if hash1 != hash2 {
		t.Errorf("hashContent should be deterministic: hash1=%s, hash2=%s", hash1, hash2)
	}

	// Different content should produce different hash
	differentContent := "different test content"
	hash3 := hashContent(differentContent)

	if hash1 == hash3 {
		t.Error("Different content should produce different hashes")
	}

	// Hash should be non-empty hex string
	if len(hash1) != 64 { // SHA256 produces 64 hex characters
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

func TestMergeAndDeduplicateDocs_ZeroMaxDocs(t *testing.T) {
	t.Parallel()

	docs := []schema.Document{
		{PageContent: "content1", Score: 0.9, Metadata: map[string]any{}},
		{PageContent: "content2", Score: 0.8, Metadata: map[string]any{}},
	}

	result := MergeAndDeduplicateDocs(docs, 0)

	if len(result) != 0 {
		t.Errorf("With maxDocs=0, expected empty result, got %d documents", len(result))
	}
}

func TestMergeAndDeduplicateDocs_ComplexScenario(t *testing.T) {
	t.Parallel()

	// Simulate multiple queries with overlapping results
	docs := []schema.Document{
		// From query 1
		{PageContent: "result A", Score: 0.85, Metadata: map[string]any{"query": 1}},
		{PageContent: "result B", Score: 0.75, Metadata: map[string]any{"query": 1}},
		{PageContent: "result C", Score: 0.65, Metadata: map[string]any{"query": 1}},

		// From query 2 (some overlap)
		{PageContent: "result A", Score: 0.90, Metadata: map[string]any{"query": 2}}, // Duplicate with higher score
		{PageContent: "result D", Score: 0.80, Metadata: map[string]any{"query": 2}},
		{PageContent: "result E", Score: 0.70, Metadata: map[string]any{"query": 2}},

		// From query 3 (some overlap)
		{PageContent: "result B", Score: 0.60, Metadata: map[string]any{"query": 3}}, // Duplicate with lower score
		{PageContent: "result F", Score: 0.88, Metadata: map[string]any{"query": 3}},
		{PageContent: "result C", Score: 0.72, Metadata: map[string]any{"query": 3}}, // Duplicate with higher score
	}

	result := MergeAndDeduplicateDocs(docs, 5)

	// Should have at most 5 unique documents
	if len(result) > 5 {
		t.Errorf("Expected at most 5 documents, got %d", len(result))
	}

	// Verify deduplication and score selection
	contentToMaxScore := map[string]float32{
		"result A": 0.90, // Max from query 2
		"result B": 0.75, // Max from query 1
		"result C": 0.72, // Max from query 3
		"result D": 0.80,
		"result E": 0.70,
		"result F": 0.88,
	}

	for _, doc := range result {
		expectedScore, exists := contentToMaxScore[doc.PageContent]
		if !exists {
			t.Errorf("Unexpected document content: %s", doc.PageContent)
			continue
		}
		if doc.Score != expectedScore {
			t.Errorf("Document '%s' has score %f, expected %f (max score)",
				doc.PageContent, doc.Score, expectedScore)
		}
	}

	// Verify sorting (top 5 by score should be: F(0.88), A(0.90), D(0.80), B(0.75), C(0.72))
	// After sorting descending: A(0.90), F(0.88), D(0.80), B(0.75), C(0.72)
	expectedOrder := []struct {
		content string
		score   float32
	}{
		{"result A", 0.90},
		{"result F", 0.88},
		{"result D", 0.80},
		{"result B", 0.75},
		{"result C", 0.72},
	}

	for i, expected := range expectedOrder {
		if result[i].PageContent != expected.content {
			t.Errorf("result[%d] content = %s, want %s", i, result[i].PageContent, expected.content)
		}
		if result[i].Score != expected.score {
			t.Errorf("result[%d] score = %f, want %f", i, result[i].Score, expected.score)
		}
	}
}
