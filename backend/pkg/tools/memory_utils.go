package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/vxcontrol/langchaingo/schema"
)

// MergeAndDeduplicateDocs merges multiple document slices, removes duplicates based on content hash,
// sorts by score in descending order, and limits the result to maxDocs.
// When duplicates are found (same PageContent), the document with the highest Score is kept.
//
// Parameters:
//   - docs: slice of documents from multiple queries
//   - maxDocs: maximum number of documents to return
//
// Returns: deduplicated and sorted slice of documents, limited to maxDocs
func MergeAndDeduplicateDocs(docs []schema.Document, maxDocs int) []schema.Document {
	if len(docs) == 0 {
		return []schema.Document{}
	}

	// Use map for deduplication: hash -> document with max score
	docMap := make(map[string]schema.Document)

	for _, doc := range docs {
		hash := hashContent(doc.PageContent)
		
		// If document with this hash already exists, keep the one with higher score
		if existing, found := docMap[hash]; found {
			if doc.Score > existing.Score {
				docMap[hash] = doc
			}
		} else {
			docMap[hash] = doc
		}
	}

	// Convert map to slice
	result := make([]schema.Document, 0, len(docMap))
	for _, doc := range docMap {
		result = append(result, doc)
	}

	// Sort by score in descending order (highest score first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	// Limit to maxDocs
	if len(result) > maxDocs {
		result = result[:maxDocs]
	}

	return result
}

// hashContent creates a deterministic SHA256 hash from document content
func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
