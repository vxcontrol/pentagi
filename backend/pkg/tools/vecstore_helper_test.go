package tools

import (
	"errors"
	"testing"
)

func TestVecstoreHelper_StoreDocumentWithEmbeddingLimit_EmbedsTheShortTextAndStoresTheFullOne(t *testing.T) {
	embedder := &fakeEmbedder{}
	db := &recordingKnowledgeDB{}

	id, err := storeDocumentWithEmbeddingLimit(t.Context(), db, embedder, "The ans", "The answer body.",
		map[string]any{"doc_type": "answer", "question": "q"})
	if err != nil {
		t.Fatalf("storing failed: %v", err)
	}

	if id != "doc-1" {
		t.Errorf("returned id %q, not the one the database assigned", id)
	}
	if len(embedder.embedded) != 1 || embedder.embedded[0] != "The ans" {
		t.Errorf("the embedder was sent %q, not the embedding text alone", embedder.embedded)
	}
	if db.inserts != 1 || db.document != "The answer body." {
		t.Errorf("%d inserts, the last storing %q, want one storing the full text", db.inserts, db.document)
	}
	if db.embedding != "[0.5,-1,0.25]" {
		t.Errorf("stored the vector as %v, not as a pgvector literal", db.embedding)
	}
	if db.meta["doc_type"] != "answer" || db.meta["question"] != "q" {
		t.Errorf("stored metadata %v", db.meta)
	}
}

func TestVecstoreHelper_StoreDocumentWithEmbeddingLimit_RefusesWhatItCannotEmbedOrInsert(t *testing.T) {
	for name, tc := range map[string]struct {
		embedder    *fakeEmbedder
		db          *recordingKnowledgeDB
		want        string
		wantInserts int
	}{
		"the embedder fails": {
			embedder: &fakeEmbedder{err: errors.New("401 Unauthorized")},
			db:       &recordingKnowledgeDB{},
			want:     "failed to compute embedding: 401 Unauthorized",
		},
		"the embedder returns no vector": {
			embedder: &fakeEmbedder{none: true},
			db:       &recordingKnowledgeDB{},
			want:     "embedder returned no vectors",
		},
		"the database refuses the insert": {
			embedder:    &fakeEmbedder{},
			db:          &recordingKnowledgeDB{err: errors.New("connection reset by peer")},
			want:        "failed to insert document with pre-computed embedding: connection reset by peer",
			wantInserts: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			id, err := storeDocumentWithEmbeddingLimit(t.Context(), tc.db, tc.embedder, "The ans", "The answer body.",
				map[string]any{"doc_type": "answer", "question": "q"})

			if err == nil || err.Error() != tc.want {
				t.Fatalf("got id %q and error %v, want the error %q", id, err, tc.want)
			}
			if tc.db.inserts != tc.wantInserts {
				t.Errorf("%d inserts reached the database, want %d", tc.db.inserts, tc.wantInserts)
			}
		})
	}
}
