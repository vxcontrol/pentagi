package tools

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	evidenceReceiptSchema  = "pentagi.evidence_receipt"
	evidenceReceiptVersion = 1
	evidenceReceiptFile    = "receipts.jsonl"

	evidenceReceiptStatusFinished = "finished"
	evidenceReceiptStatusFailed   = "failed"
)

type evidenceReceiptRecorder interface {
	RecordFinished(context.Context, evidenceReceiptEvent) error
	RecordFailed(context.Context, evidenceReceiptEvent) error
}

type evidenceReceiptEvent struct {
	FlowID     int64
	TaskID     *int64
	SubtaskID  *int64
	ToolcallID int64
	CallID     string
	ToolName   string
	Args       json.RawMessage
	Result     string
}

type evidenceReceipt struct {
	Schema              string    `json:"schema"`
	Version             int       `json:"version"`
	ReceiptID           string    `json:"receipt_id"`
	ReceiptHash         string    `json:"receipt_hash"`
	PreviousReceiptHash string    `json:"previous_receipt_hash"`
	FlowID              int64     `json:"flow_id"`
	TaskID              *int64    `json:"task_id,omitempty"`
	SubtaskID           *int64    `json:"subtask_id,omitempty"`
	ToolcallID          int64     `json:"toolcall_id"`
	CallID              string    `json:"call_id"`
	ToolName            string    `json:"tool_name"`
	Status              string    `json:"status"`
	ArgsHash            string    `json:"args_hash"`
	ResultHash          string    `json:"result_hash"`
	CreatedAt           time.Time `json:"created_at"`
}

type evidenceReceiptPayload struct {
	Schema              string    `json:"schema"`
	Version             int       `json:"version"`
	ReceiptID           string    `json:"receipt_id"`
	PreviousReceiptHash string    `json:"previous_receipt_hash"`
	FlowID              int64     `json:"flow_id"`
	TaskID              *int64    `json:"task_id,omitempty"`
	SubtaskID           *int64    `json:"subtask_id,omitempty"`
	ToolcallID          int64     `json:"toolcall_id"`
	CallID              string    `json:"call_id"`
	ToolName            string    `json:"tool_name"`
	Status              string    `json:"status"`
	ArgsHash            string    `json:"args_hash"`
	ResultHash          string    `json:"result_hash"`
	CreatedAt           time.Time `json:"created_at"`
}

type noopEvidenceReceiptRecorder struct{}

func (noopEvidenceReceiptRecorder) RecordFinished(context.Context, evidenceReceiptEvent) error {
	return nil
}

func (noopEvidenceReceiptRecorder) RecordFailed(context.Context, evidenceReceiptEvent) error {
	return nil
}

type fileEvidenceReceiptRecorder struct {
	dataDir string
	flowID  int64
	now     func() time.Time
	newID   func() string
}

var evidenceReceiptLocks sync.Map

func newEvidenceReceiptRecorder(dataDir string, flowID int64, enabled bool) evidenceReceiptRecorder {
	if !enabled {
		return noopEvidenceReceiptRecorder{}
	}

	return &fileEvidenceReceiptRecorder{
		dataDir: dataDir,
		flowID:  flowID,
		now:     time.Now,
		newID:   newEvidenceReceiptID,
	}
}

func newEvidenceReceiptID() string {
	return "receipt_" + uuid.NewString()
}

func (r *fileEvidenceReceiptRecorder) RecordFinished(ctx context.Context, event evidenceReceiptEvent) error {
	return r.record(ctx, event, evidenceReceiptStatusFinished)
}

func (r *fileEvidenceReceiptRecorder) RecordFailed(ctx context.Context, event evidenceReceiptEvent) error {
	return r.record(ctx, event, evidenceReceiptStatusFailed)
}

func (r *fileEvidenceReceiptRecorder) record(ctx context.Context, event evidenceReceiptEvent, status string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	event.FlowID = r.flowID
	path, err := evidenceReceiptsPath(r.dataDir, event.FlowID)
	if err != nil {
		return err
	}

	lock := evidenceReceiptPathLock(path)
	lock.Lock()
	defer lock.Unlock()

	previousHash, err := readLastEvidenceReceiptHash(path)
	if err != nil {
		return err
	}

	argsHash, err := hashCanonicalJSON(event.Args)
	if err != nil {
		return fmt.Errorf("failed to hash evidence receipt arguments: %w", err)
	}

	receipt := evidenceReceipt{
		Schema:              evidenceReceiptSchema,
		Version:             evidenceReceiptVersion,
		ReceiptID:           r.newID(),
		PreviousReceiptHash: previousHash,
		FlowID:              event.FlowID,
		TaskID:              event.TaskID,
		SubtaskID:           event.SubtaskID,
		ToolcallID:          event.ToolcallID,
		CallID:              event.CallID,
		ToolName:            event.ToolName,
		Status:              status,
		ArgsHash:            argsHash,
		ResultHash:          hashBytes([]byte(event.Result)),
		CreatedAt:           r.now().UTC(),
	}

	receiptHash, err := computeEvidenceReceiptHash(receipt)
	if err != nil {
		return err
	}
	receipt.ReceiptHash = receiptHash

	line, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence receipt: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to prepare evidence receipt directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open evidence receipt file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("failed to write evidence receipt: %w", err)
	}

	return nil
}

func evidenceReceiptsPath(dataDir string, flowID int64) (string, error) {
	baseDir, err := filepath.Abs(dataDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve evidence receipt data directory: %w", err)
	}

	path := filepath.Join(baseDir, fmt.Sprintf("flow-%d", flowID), "evidence", evidenceReceiptFile)
	cleanPath := filepath.Clean(path)
	rel, err := filepath.Rel(baseDir, cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to validate evidence receipt path: %w", err)
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("evidence receipt path escapes data directory")
	}

	return cleanPath, nil
}

func evidenceReceiptPathLock(path string) *sync.Mutex {
	lock, _ := evidenceReceiptLocks.LoadOrStore(path, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func readLastEvidenceReceiptHash(path string) (string, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to read evidence receipt file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	previousHash := ""
	lineNumber := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		lineNumber++

		var receipt evidenceReceipt
		if err := json.Unmarshal(line, &receipt); err != nil {
			return "", fmt.Errorf("failed to parse evidence receipt line %d: %w", lineNumber, err)
		}
		if receipt.Schema != evidenceReceiptSchema || receipt.Version != evidenceReceiptVersion {
			return "", fmt.Errorf("unsupported evidence receipt schema or version on line %d", lineNumber)
		}
		if receipt.PreviousReceiptHash != previousHash {
			return "", fmt.Errorf("evidence receipt chain mismatch on line %d", lineNumber)
		}

		hash, err := computeEvidenceReceiptHash(receipt)
		if err != nil {
			return "", fmt.Errorf("failed to hash evidence receipt line %d: %w", lineNumber, err)
		}
		if receipt.ReceiptHash != hash {
			return "", fmt.Errorf("evidence receipt hash mismatch on line %d", lineNumber)
		}

		previousHash = receipt.ReceiptHash
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to scan evidence receipt file: %w", err)
	}

	return previousHash, nil
}

func computeEvidenceReceiptHash(receipt evidenceReceipt) (string, error) {
	payload, err := json.Marshal(evidenceReceiptPayload{
		Schema:              receipt.Schema,
		Version:             receipt.Version,
		ReceiptID:           receipt.ReceiptID,
		PreviousReceiptHash: receipt.PreviousReceiptHash,
		FlowID:              receipt.FlowID,
		TaskID:              receipt.TaskID,
		SubtaskID:           receipt.SubtaskID,
		ToolcallID:          receipt.ToolcallID,
		CallID:              receipt.CallID,
		ToolName:            receipt.ToolName,
		Status:              receipt.Status,
		ArgsHash:            receipt.ArgsHash,
		ResultHash:          receipt.ResultHash,
		CreatedAt:           receipt.CreatedAt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal evidence receipt payload: %w", err)
	}

	return hashBytes(payload), nil
}

func hashCanonicalJSON(raw json.RawMessage) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return "", fmt.Errorf("unexpected extra JSON data")
	}

	canonical, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return hashBytes(canonical), nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
