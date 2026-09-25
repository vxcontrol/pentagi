package main

import (
	"errors"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/report.golden from the current output")

var generatedLine = regexp.MustCompile(`(?m)^Generated: .*$`)

func sampleResults() []AgentTestResult {
	return []AgentTestResult{
		{
			AgentType: "assistant",
			ModelName: "gpt-4o",
			Reasoning: true,
			BasicTests: []TestResult{
				{Name: "simple call", Success: true, LatencyMs: 1200},
				{Name: "streamed call", Success: true, Streaming: true, LatencyMs: 900},
				{Name: "bad call", Error: errors.New("rate limited | try again"), LatencyMs: 300},
			},
			AdvancedTests: []TestResult{
				{Name: "tool call", Success: true, LatencyMs: 2500},
				{Name: "refused call", Error: errors.New("the model returned no answer"), StopReason: "content_filter", ContentFiltered: true, LatencyMs: 400},
			},
			CapabilityTests: []TestResult{
				{Name: "adaptive thinking", Capability: "reasoning", Success: true, LatencyMs: 1500},
				{Name: "structured output", Capability: "schema", Unsupported: true, Error: errors.New("not supported")},
				{Name: "reasoning off", Capability: "reasoning", Error: errors.New("gateway timeout"), LatencyMs: 800},
			},
			TotalSuccess:   4,
			TotalTests:     5,
			TotalFiltered:  1,
			AverageLatency: 1400 * time.Millisecond,
		},
		{
			AgentType:       "installer",
			ModelName:       "claude-sonnet",
			BasicTests:      []TestResult{{Name: "simple call", Success: true, LatencyMs: 700}},
			TotalSuccess:    1,
			TotalTests:      1,
			AverageLatency:  700 * time.Millisecond,
			SkippedAdvanced: true,
			SkippedReason:   "advanced tests skipped for this model",
		},
		{
			AgentType: "empty",
			ModelName: "none",
		},
	}
}

func TestReport_WriteReportToFile_KeepsTheReportShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.md")

	if err := WriteReportToFile(sampleResults(), path); err != nil {
		t.Fatalf("WriteReportToFile: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	got := generatedLine.ReplaceAll(written, []byte("Generated: <stamped>"))
	golden := filepath.Join("testdata", "report.golden")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}

		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("report changed shape;\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestReport_WriteReportToFile_FailsForAPathThatCannotBeCreated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such-dir", "report.md")

	if err := WriteReportToFile(sampleResults(), path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("writing into a missing directory returned %v, want a not-exist error", err)
	}
}

type shortWriter struct {
	limit   int
	written int
	closed  bool
}

func (w *shortWriter) Write(p []byte) (int, error) {
	room := w.limit - w.written
	if room <= 0 {
		return 0, errors.New("no space left on device")
	}

	if len(p) <= room {
		w.written += len(p)

		return len(p), nil
	}

	w.written = w.limit

	return room, errors.New("no space left on device")
}

func (w *shortWriter) Close() error {
	w.closed = true

	return nil
}

type failingCloser struct {
	io.Writer
}

func (failingCloser) Close() error {
	return errors.New("flush on close failed")
}

type flakyWriter struct {
	failOn int
	calls  int
}

func (w *flakyWriter) Write(p []byte) (int, error) {
	w.calls++

	if w.calls == w.failOn {
		return 0, errors.New("device busy")
	}

	return len(p), nil
}

func (w *flakyWriter) Close() error {
	return nil
}

func TestReport_WriteReport_FailsWhenTheDestinationFails(t *testing.T) {
	short := &shortWriter{limit: 200}

	tests := []struct {
		name    string
		dst     io.WriteCloser
		wantErr string
		after   func(t *testing.T)
	}{
		{
			name:    "a destination that runs out of room",
			dst:     short,
			wantErr: "no space left on device",
			after: func(t *testing.T) {
				if short.written != 200 {
					t.Errorf("wrote %d bytes, expected the writer to stop at its limit", short.written)
				}
				if !short.closed {
					t.Error("the destination was not closed")
				}
			},
		},
		{name: "a close that fails", dst: failingCloser{Writer: io.Discard}, wantErr: "flush on close failed"},
		{name: "a write that fails and then recovers", dst: &flakyWriter{failOn: 3}, wantErr: "device busy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeReport(tt.dst, sampleResults())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("a report that was not fully written returned %v, want an error naming %q", err, tt.wantErr)
			}
			if tt.after != nil {
				tt.after(t)
			}
		})
	}
}

func TestReport_FailureText_NamesWhyTheModelStopped(t *testing.T) {
	for _, tc := range []struct {
		label string
		test  TestResult
		want  string
	}{
		{"error and stop reason", TestResult{Error: errors.New("no answer"), StopReason: "content_filter"},
			"no answer; stop reason: content_filter"},
		{"stop reason alone", TestResult{StopReason: "refusal"}, "stop reason: refusal"},
		{"an error that already names it", TestResult{Error: errors.New(`stop reason "length"`), StopReason: "length"},
			`stop reason "length"`},
		{"neither", TestResult{}, ""},
	} {
		t.Run(tc.label, func(t *testing.T) {
			if got := failureText(tc.test); got != tc.want {
				t.Errorf("failureText = %q, want %q", got, tc.want)
			}
		})
	}
}

type bufferCloser struct{ strings.Builder }

func (*bufferCloser) Close() error { return nil }

func TestReport_RowStatus_ShowsAnUnsupportedCaseWithItsReason(t *testing.T) {
	results := []AgentTestResult{{
		AgentType: "coder",
		ModelName: "claude-haiku-4-5",
		BasicTests: []TestResult{{
			Name:        "Answer Stops At The Output Limit",
			Success:     true,
			Unsupported: true,
			Error:       errors.New("the output limit the case asks for did not reach the vendor"),
		}},
	}}

	var out bufferCloser
	if err := writeReport(&out, results); err != nil {
		t.Fatalf("writeReport: %v", err)
	}

	if !strings.Contains(out.String(), "| Answer Stops At The Output Limit | ⊘ Not Supported |") {
		t.Errorf("the case must read as not supported, not as a pass:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "did not reach the vendor") {
		t.Errorf("the report must carry the reason:\n%s", out.String())
	}
}
