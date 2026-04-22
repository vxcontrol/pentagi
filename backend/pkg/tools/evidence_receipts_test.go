package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvidenceReceiptHashDeterministic(t *testing.T) {
	t.Parallel()

	argsHashA, err := hashCanonicalJSON(json.RawMessage(`{"b":2,"a":{"d":4,"c":3}}`))
	if err != nil {
		t.Fatalf("hashCanonicalJSON() unexpected error: %v", err)
	}
	argsHashB, err := hashCanonicalJSON(json.RawMessage(`{"a":{"c":3,"d":4},"b":2}`))
	if err != nil {
		t.Fatalf("hashCanonicalJSON() unexpected error: %v", err)
	}
	if argsHashA != argsHashB {
		t.Fatalf("canonical argument hashes differ: %s != %s", argsHashA, argsHashB)
	}

	receipt := testEvidenceReceipt(argsHashA, hashBytes([]byte("result")))
	hashA, err := computeEvidenceReceiptHash(receipt)
	if err != nil {
		t.Fatalf("computeEvidenceReceiptHash() unexpected error: %v", err)
	}
	hashB, err := computeEvidenceReceiptHash(receipt)
	if err != nil {
		t.Fatalf("computeEvidenceReceiptHash() unexpected error: %v", err)
	}
	if hashA != hashB {
		t.Fatalf("receipt hash should be deterministic: %s != %s", hashA, hashB)
	}
}

func TestEvidenceReceiptHashChangesWithContentHashes(t *testing.T) {
	t.Parallel()

	receipt := testEvidenceReceipt(hashBytes([]byte("args")), hashBytes([]byte("result")))
	baseHash, err := computeEvidenceReceiptHash(receipt)
	if err != nil {
		t.Fatalf("computeEvidenceReceiptHash() unexpected error: %v", err)
	}

	receipt.ArgsHash = hashBytes([]byte("different args"))
	argsHash, err := computeEvidenceReceiptHash(receipt)
	if err != nil {
		t.Fatalf("computeEvidenceReceiptHash() unexpected error: %v", err)
	}
	if argsHash == baseHash {
		t.Fatal("receipt hash should change when args_hash changes")
	}

	receipt = testEvidenceReceipt(hashBytes([]byte("args")), hashBytes([]byte("different result")))
	resultHash, err := computeEvidenceReceiptHash(receipt)
	if err != nil {
		t.Fatalf("computeEvidenceReceiptHash() unexpected error: %v", err)
	}
	if resultHash == baseHash {
		t.Fatal("receipt hash should change when result_hash changes")
	}
}

func TestNoopEvidenceReceiptRecorder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	recorder := newEvidenceReceiptRecorder(dir, 10, false)
	err := recorder.RecordFinished(t.Context(), testEvidenceReceiptEvent())
	if err != nil {
		t.Fatalf("RecordFinished() unexpected error: %v", err)
	}

	path, err := evidenceReceiptsPath(dir, 10)
	if err != nil {
		t.Fatalf("evidenceReceiptsPath() unexpected error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("disabled recorder should not create %s, stat err: %v", path, err)
	}
}

func TestFileEvidenceReceiptRecorderWritesJSONLAndLinksReceipts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	recorder := newTestEvidenceReceiptRecorder(dir, 42)

	if err := recorder.RecordFinished(t.Context(), testEvidenceReceiptEvent()); err != nil {
		t.Fatalf("RecordFinished() unexpected error: %v", err)
	}
	if err := recorder.RecordFinished(t.Context(), evidenceReceiptEvent{
		FlowID:     42,
		ToolcallID: 101,
		CallID:     "call-101",
		ToolName:   "terminal",
		Args:       json.RawMessage(`{"cmd":"whoami"}`),
		Result:     "root",
	}); err != nil {
		t.Fatalf("RecordFinished() second receipt unexpected error: %v", err)
	}

	path, err := evidenceReceiptsPath(dir, 42)
	if err != nil {
		t.Fatalf("evidenceReceiptsPath() unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read evidence receipts: %v", err)
	}
	if strings.Contains(string(data), "sensitive output") {
		t.Fatal("receipt file should not store raw result content")
	}
	receipts := readEvidenceReceiptLines(t, path)
	if len(receipts) != 2 {
		t.Fatalf("got %d receipts, want 2", len(receipts))
	}

	if receipts[0].Status != evidenceReceiptStatusFinished {
		t.Fatalf("first receipt status = %q, want %q", receipts[0].Status, evidenceReceiptStatusFinished)
	}
	if receipts[0].PreviousReceiptHash != "" {
		t.Fatalf("first receipt previous hash = %q, want empty", receipts[0].PreviousReceiptHash)
	}
	if receipts[1].PreviousReceiptHash != receipts[0].ReceiptHash {
		t.Fatalf("second receipt previous hash = %q, want %q", receipts[1].PreviousReceiptHash, receipts[0].ReceiptHash)
	}
}

func TestFileEvidenceReceiptRecorderRecordsFailedStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	recorder := newTestEvidenceReceiptRecorder(dir, 77)

	if err := recorder.RecordFailed(t.Context(), testEvidenceReceiptEvent()); err != nil {
		t.Fatalf("RecordFailed() unexpected error: %v", err)
	}

	path, err := evidenceReceiptsPath(dir, 77)
	if err != nil {
		t.Fatalf("evidenceReceiptsPath() unexpected error: %v", err)
	}
	receipts := readEvidenceReceiptLines(t, path)
	if len(receipts) != 1 {
		t.Fatalf("got %d receipts, want 1", len(receipts))
	}
	if receipts[0].Status != evidenceReceiptStatusFailed {
		t.Fatalf("receipt status = %q, want %q", receipts[0].Status, evidenceReceiptStatusFailed)
	}
}

func TestFileEvidenceReceiptRecorderFailsClosedOnUnreadablePreviousReceipt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path, err := evidenceReceiptsPath(dir, 88)
	if err != nil {
		t.Fatalf("evidenceReceiptsPath() unexpected error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create evidence dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"schema":"wrong"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to seed corrupt receipt: %v", err)
	}

	recorder := newTestEvidenceReceiptRecorder(dir, 88)
	err = recorder.RecordFinished(t.Context(), testEvidenceReceiptEvent())
	if err == nil {
		t.Fatal("RecordFinished() should fail when previous receipt cannot be verified")
	}
}

func testEvidenceReceipt(argsHash, resultHash string) evidenceReceipt {
	return evidenceReceipt{
		Schema:              evidenceReceiptSchema,
		Version:             evidenceReceiptVersion,
		ReceiptID:           "receipt_test",
		PreviousReceiptHash: "sha256:previous",
		FlowID:              42,
		ToolcallID:          100,
		CallID:              "call-100",
		ToolName:            "terminal",
		Status:              evidenceReceiptStatusFinished,
		ArgsHash:            argsHash,
		ResultHash:          resultHash,
		CreatedAt:           time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),
	}
}

func testEvidenceReceiptEvent() evidenceReceiptEvent {
	taskID := int64(20)
	subtaskID := int64(30)

	return evidenceReceiptEvent{
		FlowID:     10,
		TaskID:     &taskID,
		SubtaskID:  &subtaskID,
		ToolcallID: 100,
		CallID:     "call-100",
		ToolName:   "terminal",
		Args:       json.RawMessage(`{"cmd":"id"}`),
		Result:     "sensitive output",
	}
}

func newTestEvidenceReceiptRecorder(dir string, flowID int64) *fileEvidenceReceiptRecorder {
	var count int

	return &fileEvidenceReceiptRecorder{
		dataDir: dir,
		flowID:  flowID,
		now: func() time.Time {
			count++
			return time.Date(2026, 4, 22, 12, 0, count, 0, time.UTC)
		},
		newID: func() string {
			return fmt.Sprintf("receipt_%03d", count+1)
		},
	}
}

func readEvidenceReceiptLines(t *testing.T, path string) []evidenceReceipt {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read evidence receipts: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	receipts := make([]evidenceReceipt, 0, len(lines))
	for _, line := range lines {
		var receipt evidenceReceipt
		if err := json.Unmarshal([]byte(line), &receipt); err != nil {
			t.Fatalf("failed to parse receipt line %q: %v", line, err)
		}
		receipts = append(receipts, receipt)
	}

	return receipts
}
