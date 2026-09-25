package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"pentagi/pkg/graph/model"
)

func boolPtr(b bool) *Bool {
	v := Bool(b)
	return &v
}

func int64Ptr(i int64) *Int64 {
	v := Int64(i)
	return &v
}

// Decodes whole action structs, so each lenient type is proven wired to the field a model fills.
func TestArgs_AcceptsTheSerialisationSlipsModelsMake(t *testing.T) {
	t.Parallel()

	const doubleEncodedQuestions = `"[\"SCCM MECM CM_P01 database SC_NAA Network Access Account credentials extraction\", \"SCCM ClientKeyData RSA private keys mTLS bypass MECM\"]"`
	decodedQuestions := Strings{
		"SCCM MECM CM_P01 database SC_NAA Network Access Account credentials extraction",
		"SCCM ClientKeyData RSA private keys mTLS bypass MECM",
	}

	tests := []struct {
		name string
		args string
		into any
		want any
	}{
		{
			name: "file action and path wrapped in an extra pair of quotes",
			args: `{"action": "\"write_file\"", "path": "\"/home/evidence/inject.py\"", "content": "print(1)", "message": "m"}`,
			into: &FileAction{},
			want: &FileAction{Action: WriteFile, Path: "/home/evidence/inject.py", Content: "print(1)", Message: "m"},
		},
		{
			name: "read_file action wrapped in an extra pair of quotes",
			args: `{"action": "\"read_file\"", "path": "/home/evidence/inject.py", "message": "m"}`,
			into: &FileAction{},
			want: &FileAction{Action: ReadFile, Path: "/home/evidence/inject.py", Message: "m"},
		},
		{
			name: "browser action wrapped in an extra pair of quotes",
			args: `{"url": "http://10.0.0.5/", "action": "\"markdown\"", "message": "m"}`,
			into: &Browser{},
			want: &Browser{Url: "http://10.0.0.5/", Action: Markdown, Message: "m"},
		},
		{
			name: "subtask operation wrapped in an extra pair of quotes",
			args: `{"op": "\"add\"", "title": "t", "description": "d"}`,
			into: &SubtaskOperation{},
			want: &SubtaskOperation{Op: SubtaskOpAdd, Title: "t", Description: "d"},
		},
		{
			name: "flow status detail wrapped in an extra pair of quotes",
			args: `{"detail": "\"summary\"", "message": "m"}`,
			into: &GetFlowStatusAction{},
			want: &GetFlowStatusAction{Detail: FlowStatusDetailSummary, Message: "m"},
		},
		{
			name: "terminal timeout sent as an empty string",
			args: `{"input": "curl -s http://example.com/", "cwd": "/", "detach": false, "message": "test", "timeout": ""}`,
			into: &TerminalAction{},
			want: &TerminalAction{Input: "curl -s http://example.com/", Cwd: "/", Message: "test"},
		},
		{
			name: "terminal detach and timeout sent as strings",
			args: `{"input": "nmap -sV 10.0.0.5", "cwd": "/work", "detach": "true", "timeout": "30", "message": "m"}`,
			into: &TerminalAction{},
			want: &TerminalAction{Input: "nmap -sV 10.0.0.5", Cwd: "/work", Detach: true, Timeout: 30, Message: "m"},
		},
		{
			name: "search_in_memory questions as a real array",
			args: `{"questions": ["q1", "q2"], "message": "m"}`,
			into: &SearchInMemoryAction{},
			want: &SearchInMemoryAction{Questions: Strings{"q1", "q2"}, Message: "m"},
		},
		{
			name: "search_in_memory questions double-encoded",
			args: `{"max_results": 10, "message": "m", "questions": ` + doubleEncodedQuestions + `}`,
			into: &SearchInMemoryAction{},
			want: &SearchInMemoryAction{Questions: decodedQuestions, Message: "m"},
		},
		{
			name: "search_guide questions as a real array",
			args: `{"questions": ["q1", "q2"], "type": "install", "message": "m"}`,
			into: &SearchGuideAction{},
			want: &SearchGuideAction{Questions: Strings{"q1", "q2"}, Type: GuideType(model.KnowledgeGuideTypeInstall), Message: "m"},
		},
		{
			name: "search_guide questions double-encoded",
			args: `{"questions": ` + doubleEncodedQuestions + `, "type": "pentest", "message": "m"}`,
			into: &SearchGuideAction{},
			want: &SearchGuideAction{Questions: decodedQuestions, Type: GuideType(model.KnowledgeGuideTypePentest), Message: "m"},
		},
		{
			name: "search_answer questions as a real array",
			args: `{"questions": ["q1", "q2"], "type": "tool", "message": "m"}`,
			into: &SearchAnswerAction{},
			want: &SearchAnswerAction{Questions: Strings{"q1", "q2"}, Type: AnswerType(model.KnowledgeAnswerTypeTool), Message: "m"},
		},
		{
			name: "search_answer questions double-encoded",
			args: `{"questions": ` + doubleEncodedQuestions + `, "type": "vulnerability", "message": "m"}`,
			into: &SearchAnswerAction{},
			want: &SearchAnswerAction{Questions: decodedQuestions, Type: AnswerType(model.KnowledgeAnswerTypeVulnerability), Message: "m"},
		},
		{
			name: "search_code questions as a real array",
			args: `{"questions": ["q1", "q2"], "lang": "python", "message": "m"}`,
			into: &SearchCodeAction{},
			want: &SearchCodeAction{Questions: Strings{"q1", "q2"}, Lang: "python", Message: "m"},
		},
		{
			name: "search_code questions double-encoded",
			args: `{"questions": ` + doubleEncodedQuestions + `, "lang": "bash", "message": "m"}`,
			into: &SearchCodeAction{},
			want: &SearchCodeAction{Questions: decodedQuestions, Lang: "bash", Message: "m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := json.Unmarshal([]byte(tt.args), tt.into); err != nil {
				t.Fatalf("the arguments were refused: %v", err)
			}
			if !reflect.DeepEqual(tt.into, tt.want) {
				t.Errorf("decoded %+v, want %+v", tt.into, tt.want)
			}
		})
	}
}

// Recovered only when the message field is empty and nothing but whitespace separates the two tags.
func TestArgs_SplitLeakedMessage_RecoversAMessageWrittenBehindAFieldBoundary(t *testing.T) {
	results := []struct {
		name   string
		decode func(args []byte) (result, message string, err error)
	}{
		{"done", func(args []byte) (string, string, error) {
			var v Done
			err := json.Unmarshal(args, &v)
			return v.Result, v.Message, err
		}},
		{"task result", func(args []byte) (string, string, error) {
			var v TaskResult
			err := json.Unmarshal(args, &v)
			return v.Result, v.Message, err
		}},
		{"search result", func(args []byte) (string, string, error) {
			var v SearchResult
			err := json.Unmarshal(args, &v)
			return v.Result, v.Message, err
		}},
	}

	tests := map[string]struct {
		args        string
		wantResult  string
		wantMessage string
		wantErr     bool
	}{
		"the model closed the result and opened the message inside it": {
			args:        `{"success": true, "result": "The scan finished.\n</parameter>\n<parameter name=\"message\">\n  Scan done, two ports open.\n", "message": ""}`,
			wantResult:  "The scan finished.",
			wantMessage: "Scan done, two ports open.",
		},
		"the result quotes a closing tag with text before a message tag": {
			args:        `{"success": true, "result": "Body: </parameter> then <parameter name=\"message\">hi", "message": ""}`,
			wantResult:  "Body: </parameter> then <parameter name=\"message\">hi",
			wantMessage: "",
		},
		"the model sent both fields": {
			args:        `{"success": true, "result": "The scan finished.\n</parameter>\n<parameter name=\"message\">stray", "message": "Scan done."}`,
			wantResult:  "The scan finished.\n</parameter>\n<parameter name=\"message\">stray",
			wantMessage: "Scan done.",
		},
		"the result quotes the markup of a scanned target": {
			args:        `{"success": true, "result": "The page contains <parameter name=\"message\">hello</parameter> in its body.", "message": ""}`,
			wantResult:  "The page contains <parameter name=\"message\">hello</parameter> in its body.",
			wantMessage: "",
		},
		"the result carries no markup": {
			args:        `{"success": true, "result": "The scan finished.", "message": "Scan done."}`,
			wantResult:  "The scan finished.",
			wantMessage: "Scan done.",
		},
		"the model sent no message and no markup": {
			args:        `{"success": true, "result": "The scan finished.", "message": ""}`,
			wantResult:  "The scan finished.",
			wantMessage: "",
		},
		"a wrongly typed field is refused, not decoded half-empty": {
			args:    `{"success": true, "result": 5, "message": ""}`,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, r := range results {
				result, message, err := r.decode([]byte(tc.args))
				if tc.wantErr {
					if err == nil {
						t.Errorf("%s: decoded result %q, want a refusal", r.name, result)
					}
					continue
				}
				if err != nil {
					t.Fatalf("unmarshal %s: %v", r.name, err)
				}
				if result != tc.wantResult {
					t.Errorf("%s result = %q, want %q", r.name, result, tc.wantResult)
				}
				if message != tc.wantMessage {
					t.Errorf("%s message = %q, want %q", r.name, message, tc.wantMessage)
				}
			}
		})
	}
}

func TestArgs_SubtaskPatch_NamesTheOperationItRefusesAndWhy(t *testing.T) {
	t.Parallel()

	id := int64(1)

	tests := []struct {
		name    string
		ops     []SubtaskOperation
		wantErr string
	}{
		{
			name: "valid add operation",
			ops:  []SubtaskOperation{{Op: SubtaskOpAdd, Title: "t", Description: "d"}},
		},
		{
			name:    "add missing title",
			ops:     []SubtaskOperation{{Op: SubtaskOpAdd, Description: "d"}},
			wantErr: "operation 0: add requires title",
		},
		{
			name:    "add missing description",
			ops:     []SubtaskOperation{{Op: SubtaskOpAdd, Title: "t"}},
			wantErr: "operation 0: add requires description",
		},
		{
			name: "remove with id",
			ops:  []SubtaskOperation{{Op: SubtaskOpRemove, ID: &id}},
		},
		{
			name:    "remove without id",
			ops:     []SubtaskOperation{{Op: SubtaskOpRemove}},
			wantErr: "operation 0: remove requires id",
		},
		{
			name: "modify with id and title",
			ops:  []SubtaskOperation{{Op: SubtaskOpModify, ID: &id, Title: "new title"}},
		},
		{
			name: "modify with id and only a description",
			ops:  []SubtaskOperation{{Op: SubtaskOpModify, ID: &id, Description: "new description"}},
		},
		{
			name:    "modify without id",
			ops:     []SubtaskOperation{{Op: SubtaskOpModify, Title: "t"}},
			wantErr: "operation 0: modify requires id",
		},
		{
			name:    "modify without title or description",
			ops:     []SubtaskOperation{{Op: SubtaskOpModify, ID: &id}},
			wantErr: "operation 0: modify requires at least title or description",
		},
		{
			name: "reorder with id",
			ops:  []SubtaskOperation{{Op: SubtaskOpReorder, ID: &id}},
		},
		{
			name:    "reorder without id",
			ops:     []SubtaskOperation{{Op: SubtaskOpReorder}},
			wantErr: "operation 0: reorder requires id",
		},
		{
			name:    "unknown operation type",
			ops:     []SubtaskOperation{{Op: "unknown"}},
			wantErr: `operation 0: unknown operation type "unknown"`,
		},
		{
			name: "empty operations list is valid",
			ops:  []SubtaskOperation{},
		},
		{
			name: "the failing operation is named by its position",
			ops: []SubtaskOperation{
				{Op: SubtaskOpAdd, Title: "t", Description: "d"},
				{Op: SubtaskOpRemove},
			},
			wantErr: "operation 1: remove requires id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := SubtaskPatch{Operations: tt.ops}.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("Validate() = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestArgs_SubtaskInfos_AcceptsADoubleEncodedList(t *testing.T) {
	t.Parallel()

	want := SubtaskList{
		Subtasks: SubtaskInfos{{Title: "recon", Description: "map the surface"}},
		Message:  "planned",
	}

	tests := []struct {
		name string
		args string
	}{
		{
			name: "a plan double-encoded as a string",
			args: `{"subtasks":"[{\"title\":\"recon\",\"description\":\"map the surface\"}]","message":"planned"}`,
		},
		{
			name: "a plain array",
			args: `{"subtasks":[{"title":"recon","description":"map the surface"}],"message":"planned"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got SubtaskList
			if err := json.Unmarshal([]byte(tt.args), &got); err != nil {
				t.Fatalf("the plan was refused: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decoded %+v, want %+v", got, want)
			}
		})
	}
}

func TestArgs_SubtaskInfos_RefusesWhatItCannotRecover(t *testing.T) {
	t.Parallel()

	for _, tt := range argsListRefusals("subtasks", `[{"title": 5}]`, "SubtaskInfo", "title") {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var infos SubtaskInfos
			err := json.Unmarshal([]byte(tt.args), &infos)
			tt.check(t, err)
		})
	}
}

func TestArgs_SubtaskOperations_AcceptsADoubleEncodedList(t *testing.T) {
	t.Parallel()

	seven := int64(7)

	tests := []struct {
		name string
		args string
		want SubtaskPatch
	}{
		{
			name: "a batch double-encoded as a string",
			args: `{"operations":"[{\"op\":\"add\",\"title\":\"scan\",\"description\":\"run nmap\"}]","message":"one step"}`,
			want: SubtaskPatch{
				Operations: SubtaskOperations{{Op: SubtaskOpAdd, Title: "scan", Description: "run nmap"}},
				Message:    "one step",
			},
		},
		{
			name: "a plain array",
			args: `{"operations":[{"op":"remove","id":7}],"message":"drop it"}`,
			want: SubtaskPatch{Operations: SubtaskOperations{{Op: SubtaskOpRemove, ID: &seven}}, Message: "drop it"},
		},
		{
			name: "an empty array",
			args: `{"operations":[],"message":"nothing to change"}`,
			want: SubtaskPatch{Operations: SubtaskOperations{}, Message: "nothing to change"},
		},
		{
			name: "an empty array double-encoded",
			args: `{"operations":"[]","message":"nothing to change"}`,
			want: SubtaskPatch{Operations: SubtaskOperations{}, Message: "nothing to change"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got SubtaskPatch
			if err := json.Unmarshal([]byte(tt.args), &got); err != nil {
				t.Fatalf("the patch was refused: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decoded %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestArgs_SubtaskOperations_RefusesWhatItCannotRecover(t *testing.T) {
	t.Parallel()

	tests := append(
		argsListRefusals("operations", `[{"op":"modify","id":"7"}]`, "SubtaskOperation", "id"),
		argsListRefusal{
			name: "a scalar",
			args: `42`,
			want: []string{"cannot unmarshal number"},
		},
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ops SubtaskOperations
			err := json.Unmarshal([]byte(tt.args), &ops)
			tt.check(t, err)
		})
	}
}

// argsListRefusal is a payload a lenient subtask list must refuse, and the words its refusal must carry.
type argsListRefusal struct {
	name    string
	args    string
	wrapper string // the field a DoubleEncodedListError must name; empty when none may be returned
	want    []string
}

// Shared by SubtaskInfos and SubtaskOperations; badElement's one element has a mistyped elementField.
func argsListRefusals(field, badElement, elementType, elementField string) []argsListRefusal {
	return []argsListRefusal{
		{
			name:    "a wrapper whose contents do not parse",
			args:    `"[{\"title\": broken"`,
			wrapper: field,
			want:    []string{field, "arrived as a string containing JSON"},
		},
		{
			// A real array keeps the ordinary decode error; the wrapper is only a fallback.
			name: "an array with a badly typed element",
			args: badElement,
			want: []string{elementType, elementField},
		},
		{
			// The escape repair must not reach for this: a spurious quote after an
			// unquoted number is a wrong argument, and guessing at it would invent data.
			name:    "a stray quote after a numeric id",
			args:    `"[{\"op\": \"modify\", \"id\":7\", \"title\": \"x\"}]"`,
			wrapper: field,
			want:    []string{field, "arrived as a string containing JSON"},
		},
		{
			// Anthropic markup emitted inside the value: nothing in it is a list,
			// so the refusal must ask for the right shape, not report a parse error.
			name: "the model's own tool-call markup",
			args: `"<parameter name=\"title\">Map the login endpoint"`,
			want: []string{field, "tool-call markup", "JSON array of objects"},
		},
		{
			// A null would decode to an empty list, masking a required field left out.
			name: "null",
			args: `null`,
			want: []string{field, "expected a JSON array", "null"},
		},
		{
			name: "a quoted null",
			args: `"null"`,
			want: []string{field, "expected a JSON array", "null"},
		},
	}
}

func (r argsListRefusal) check(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("the payload was accepted, want a refusal")
	}

	var wrapped *DoubleEncodedListError
	isWrapped := errors.As(err, &wrapped)
	switch {
	case r.wrapper == "" && isWrapped:
		t.Errorf("refusal %q is a DoubleEncodedListError, want the ordinary decode error", err)
	case r.wrapper != "" && !isWrapped:
		t.Fatalf("refusal %q is not a DoubleEncodedListError", err)
	case r.wrapper != "":
		if wrapped.Field != r.wrapper {
			t.Errorf("the refusal names %q, want %q", wrapped.Field, r.wrapper)
		}
		if wrapped.Unwrap() == nil {
			t.Error("the parse failure must stay reachable for the log")
		}
	}

	for _, want := range r.want {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not mention %q", err, want)
		}
	}
}

func TestArgs_LenientListsStillAdvertiseAnArray(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{}
	for _, tc := range []struct{ tool, field string }{
		{"subtask_list", "subtasks"},
		{"subtask_patch", "operations"},
		{"patch_flow_subtasks", "operations"},
	} {
		sch, err := ce.GetToolSchema(tc.tool)
		if err != nil {
			t.Fatalf("%s: %v", tc.tool, err)
		}
		blob, err := json.Marshal(sch)
		if err != nil {
			t.Fatalf("%s: %v", tc.tool, err)
		}

		var doc map[string]any
		if err := json.Unmarshal(blob, &doc); err != nil {
			t.Fatalf("%s: %v", tc.tool, err)
		}
		props, _ := doc["properties"].(map[string]any)
		field, _ := props[tc.field].(map[string]any)
		if field == nil {
			t.Fatalf("%s: no %s in the schema", tc.tool, tc.field)
		}
		if field["type"] != "array" {
			t.Errorf("%s.%s must stay an array, got %v", tc.tool, tc.field, field["type"])
		}
		if field["items"] == nil {
			t.Errorf("%s.%s lost its item schema", tc.tool, tc.field)
		}
	}
}

func TestArgs_Bool_AcceptsAQuotedOrPaddedBoolean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Bool
		wantErr bool
	}{
		{name: "bare true", input: `true`, want: true},
		{name: "bare false", input: `false`, want: false},
		{name: "quoted true", input: `"true"`, want: true},
		{name: "quoted false", input: `"false"`, want: false},
		{name: "upper TRUE", input: `"TRUE"`, want: true},
		{name: "upper FALSE", input: `"FALSE"`, want: false},
		{name: "mixed case True", input: `"True"`, want: true},
		{name: "mixed case False", input: `"False"`, want: false},
		{name: "single-quoted true", input: `"'true'"`, want: true},
		{name: "single-quoted false", input: `"'false'"`, want: false},
		{name: "whitespace padded true", input: `" true "`, want: true},
		{name: "whitespace padded false", input: `" false "`, want: false},
		{name: "tab and newline around bare true", input: "\n\ttrue\t\n", want: true},
		{name: "carriage return around bare true", input: "\rtrue\r", want: true},
		{name: "escaped whitespace string true should fail", input: `"\\ttrue\\n"`, wantErr: true},
		{name: "null literal", input: `null`, wantErr: true},
		{name: "invalid yes", input: `"yes"`, wantErr: true},
		{name: "invalid 1", input: `"1"`, wantErr: true},
		{name: "invalid 0", input: `"0"`, wantErr: true},
		{name: "empty string", input: `""`, wantErr: true},
		{name: "invalid word", input: `"maybe"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var b Bool
			err := b.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && b != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, b, tt.want)
			}
		})
	}
}

func TestArgs_Bool_ReportsItsValueAndToleratesANilPointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		b          *Bool
		wantJSON   string
		wantBool   bool
		wantString string
	}{
		{name: "true value", b: boolPtr(true), wantJSON: "true", wantBool: true, wantString: "true"},
		{name: "false value", b: boolPtr(false), wantJSON: "false", wantBool: false, wantString: "false"},
		{name: "nil pointer", b: nil, wantJSON: "false", wantBool: false, wantString: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.b.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", data, tt.wantJSON)
			}
			if got := tt.b.Bool(); got != tt.wantBool {
				t.Errorf("Bool() = %v, want %v", got, tt.wantBool)
			}
			if got := tt.b.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}

			var back Bool
			if err := back.UnmarshalJSON(data); err != nil || bool(back) != tt.wantBool {
				t.Errorf("%s decoded back as %v (error %v), want %v", data, back, err, tt.wantBool)
			}
		})
	}
}

func TestArgs_Int64_AcceptsAQuotedOrPaddedInteger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Int64
		wantErr bool
	}{
		{name: "bare positive", input: `42`, want: 42},
		{name: "bare negative", input: `-7`, want: -7},
		{name: "bare zero", input: `0`, want: 0},
		{name: "quoted positive", input: `"123"`, want: 123},
		{name: "quoted negative", input: `"-456"`, want: -456},
		{name: "quoted zero", input: `"0"`, want: 0},
		{name: "single-quoted positive", input: `"'789'"`, want: 789},
		{name: "single-quoted negative", input: `"'-5'"`, want: -5},
		{name: "whitespace padded", input: `" 100 "`, want: 100},
		{name: "tab around bare value", input: "\t99\t", want: 99},
		{name: "newline around bare value", input: "\n50\n", want: 50},
		{name: "escaped whitespace string int should fail", input: `"\\n50\\n"`, wantErr: true},
		{name: "max int64", input: `"9223372036854775807"`, want: Int64(9223372036854775807)},
		{name: "min int64", input: `"-9223372036854775808"`, want: Int64(-9223372036854775808)},
		{name: "null literal", input: `null`, wantErr: true},
		{name: "overflow int64", input: `"9223372036854775808"`, wantErr: true},
		{name: "underflow int64", input: `"-9223372036854775809"`, wantErr: true},
		{name: "invalid string", input: `"abc"`, wantErr: true},
		{name: "invalid float", input: `"1.5"`, wantErr: true},
		{name: "empty string treated as 0", input: `""`, want: 0},
		{name: "bool string", input: `"true"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var i Int64
			err := i.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && i != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, i, tt.want)
			}
		})
	}
}

func TestArgs_Int64_ReportsItsValueAndToleratesANilPointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		i          *Int64
		wantJSON   string
		wantInt    int
		wantInt64  int64
		wantPtr    bool // PtrInt64 returns non-nil
		wantString string
	}{
		{name: "positive value", i: int64Ptr(42), wantJSON: "42", wantInt: 42, wantInt64: 42, wantPtr: true, wantString: "42"},
		{name: "negative value", i: int64Ptr(-7), wantJSON: "-7", wantInt: -7, wantInt64: -7, wantPtr: true, wantString: "-7"},
		{name: "zero value", i: int64Ptr(0), wantJSON: "0", wantInt: 0, wantInt64: 0, wantPtr: true, wantString: "0"},
		{name: "nil pointer", i: nil, wantJSON: "0", wantInt: 0, wantInt64: 0, wantPtr: false, wantString: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.i.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", data, tt.wantJSON)
			}
			if got := tt.i.Int(); got != tt.wantInt {
				t.Errorf("Int() = %v, want %v", got, tt.wantInt)
			}
			if got := tt.i.Int64(); got != tt.wantInt64 {
				t.Errorf("Int64() = %v, want %v", got, tt.wantInt64)
			}
			switch got := tt.i.PtrInt64(); {
			case !tt.wantPtr && got != nil:
				t.Errorf("PtrInt64() = %v, want nil", *got)
			case tt.wantPtr && got == nil:
				t.Error("PtrInt64() = nil, want non-nil")
			case tt.wantPtr && *got != tt.wantInt64:
				t.Errorf("PtrInt64() = %v, want %v", *got, tt.wantInt64)
			}
			if got := tt.i.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}

			var back Int64
			if err := back.UnmarshalJSON(data); err != nil || int64(back) != tt.wantInt64 {
				t.Errorf("%s decoded back as %v (error %v), want %v", data, back, err, tt.wantInt64)
			}
		})
	}
}

func TestArgs_String_UnwrapsRedundantQuotes(t *testing.T) {
	t.Parallel()

	// Raw literals keep Go escape processing apart from the JSON escaping under test.
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain string", input: `"write_file"`, want: "write_file"},
		{name: "empty string", input: `""`, want: ""},
		{
			name:  "single extra-quoted",
			input: `"\"write_file\""`,
			want:  "write_file",
		},
		{
			name:  "single extra-quoted with whitespace",
			input: `"  \"read_file\"  "`,
			want:  "read_file",
		},
		{
			name:  "two extra pairs of quotes",
			input: `"\"\"markdown\"\""`,
			want:  "markdown",
		},
		{
			name:  "three extra pairs of quotes",
			input: `"\"\"\"markdown\"\"\""`,
			want:  "markdown",
		},
		{name: "number is an error", input: `42`, wantErr: true},
		{name: "null is an error", input: `null`, wantErr: true},
		{name: "bool is an error", input: `true`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s String
			err := s.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(s) != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %q, want %q", tt.input, string(s), tt.want)
			}
		})
	}
}

func TestArgs_String_MarshalsAndPrintsTheBareValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		s          String
		wantJSON   string
		wantString string
	}{
		{name: "plain value", s: String("write_file"), wantJSON: `"write_file"`, wantString: "write_file"},
		{name: "empty value", s: String(""), wantJSON: `""`, wantString: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.s.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", data, tt.wantJSON)
			}

			var stringer fmt.Stringer = tt.s
			if got := stringer.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}

			var back String
			if err := back.UnmarshalJSON(data); err != nil || string(back) != tt.wantString {
				t.Errorf("%s decoded back as %q (error %v), want %q", data, back, err, tt.wantString)
			}
		})
	}
}

func TestArgs_Strings_RecoversAListSentAsAString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "real array of strings",
			input: `["query1", "query2"]`,
			want:  []string{"query1", "query2"},
		},
		{
			name:  "real array single element",
			input: `["only one"]`,
			want:  []string{"only one"},
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  []string{},
		},
		{
			name:  "double-encoded array",
			input: `"[\"SCCM MECM CM_P01\", \"ClientKeyData RSA private keys\"]"`,
			want:  []string{"SCCM MECM CM_P01", "ClientKeyData RSA private keys"},
		},
		{
			name:  "double-encoded single-element array",
			input: `"[\"only one\"]"`,
			want:  []string{"only one"},
		},
		{
			name:  "bare string falls back to a single trimmed element",
			input: `"  just a plain question, not JSON at all\n"`,
			want:  []string{"just a plain question, not JSON at all"},
		},
		{
			name:    "empty bare string is an error",
			input:   `""`,
			wantErr: true,
		},
		{
			name:    "whitespace-only bare string is an error",
			input:   `"  \t "`,
			wantErr: true,
		},
		{
			name:    "null literal is an error",
			input:   `null`,
			wantErr: true,
		},
		{
			name:    "array of non-strings is an error",
			input:   `[1, 2, 3]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Strings
			err := s.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(s) != len(tt.want) {
				t.Fatalf("UnmarshalJSON(%s) = %v, want %v", tt.input, []string(s), tt.want)
			}
			for i := range tt.want {
				if s[i] != tt.want[i] {
					t.Errorf("UnmarshalJSON(%s)[%d] = %q, want %q", tt.input, i, s[i], tt.want[i])
				}
			}
		})
	}
}

func TestArgs_Strings_MarshalsANilListAsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    Strings
		want string
	}{
		{name: "multiple elements", s: Strings{"a", "b"}, want: `["a","b"]`},
		{name: "single element", s: Strings{"only"}, want: `["only"]`},
		{name: "empty slice", s: Strings{}, want: "[]"},
		{name: "nil slice", s: nil, want: "[]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.s.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", data, tt.want)
			}

			var back Strings
			if err := back.UnmarshalJSON(data); err != nil {
				t.Fatalf("%s did not decode back: %v", data, err)
			}
			if len(back) != len(tt.s) {
				t.Fatalf("%s decoded back as %v, want %v", data, []string(back), []string(tt.s))
			}
			for i := range tt.s {
				if back[i] != tt.s[i] {
					t.Errorf("%s decoded back as %v, want %v", data, []string(back), []string(tt.s))
				}
			}
		})
	}
}

// A type the model spelled differently must land on the enum value a search filters on.
func TestArgs_GuideType_KeepsTheDocumentFindable(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    model.KnowledgeGuideType
		wantErr bool
	}{
		{name: "already canonical", arg: `"install"`, want: model.KnowledgeGuideTypeInstall},
		{name: "the model shouted it", arg: `"Install"`, want: model.KnowledgeGuideTypeInstall},
		{name: "padded", arg: `"  use  "`, want: model.KnowledgeGuideTypeUse},
		{name: "wrapped in an extra pair of quotes", arg: `"\"Install\""`, want: model.KnowledgeGuideTypeInstall},
		{name: "outside the enum", arg: `"sorcery"`, want: model.KnowledgeGuideTypeOther},
		{name: "empty", arg: `""`, want: model.KnowledgeGuideTypeOther},
		// A null is a required field left out, not a type to file under "other".
		{name: "null", arg: `null`, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var action SearchGuideAction
			err := json.Unmarshal([]byte(`{"questions":["q"],"type":`+tc.arg+`,"message":"m"}`), &action)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("the arguments were accepted with type %q, want a refusal", action.Type)
				}
				return
			}
			if err != nil {
				t.Fatalf("the model's arguments did not parse: %v", err)
			}

			got := model.KnowledgeGuideType(action.Type)
			if got != tc.want {
				t.Errorf("stored and searched as %q, so a search for %q would miss it; wanted %q", got, tc.want, tc.want)
			}
			if !got.IsValid() {
				t.Errorf("%q is not a member of the enum, so no other door accepts it", got)
			}
		})
	}
}

// Composes both doors; CanonicalCodeLang's own classes are pinned in pkg/database/knowledge/limits.
func TestArgs_CodeLang_IsTheSameAtTheStoreAndTheSearchDoor(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    CodeLang
		wantErr bool
	}{
		{
			name: "a name carrying statement punctuation and a quote",
			arg:  `"  Objective-C'; DROP TABLE docs;--  "`,
			want: "Objective-C DROP TABLE docs--",
		},
		{name: "null is refused at both doors", arg: `null`, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var (
				search SearchCodeAction
				store  StoreCodeAction
			)
			searchErr := json.Unmarshal([]byte(`{"questions":["how to scan"],"lang":`+tc.arg+`,"message":"m"}`), &search)
			storeErr := json.Unmarshal([]byte(`{"lang":`+tc.arg+`}`), &store)
			if tc.wantErr {
				if searchErr == nil || storeErr == nil {
					t.Fatalf("search error %v, store error %v; want both doors to refuse", searchErr, storeErr)
				}
				return
			}
			if searchErr != nil || storeErr != nil {
				t.Fatalf("search error %v, store error %v", searchErr, storeErr)
			}

			if search.Lang != store.Lang {
				t.Fatalf("store wrote lang %q while search filters on %q, so the document is unreachable",
					store.Lang, search.Lang)
			}
			if search.Lang != tc.want {
				t.Errorf("lang canonicalized to %q, want %q", search.Lang, tc.want)
			}
			if strings.ContainsRune(search.Lang.String(), '\'') {
				t.Errorf("lang reaches the filter as %q, still carrying a quote", search.Lang)
			}
		})
	}
}

func TestArgs_AnswerType_KeepsTheDocumentFindable(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    model.KnowledgeAnswerType
		wantErr bool
	}{
		{name: "already canonical", arg: `"vulnerability"`, want: model.KnowledgeAnswerTypeVulnerability},
		{name: "the model shouted it", arg: `"Vulnerability"`, want: model.KnowledgeAnswerTypeVulnerability},
		{name: "padded", arg: `" tool "`, want: model.KnowledgeAnswerTypeTool},
		{name: "wrapped in an extra pair of quotes", arg: `"\"Tool\""`, want: model.KnowledgeAnswerTypeTool},
		{name: "outside the enum", arg: `"sorcery"`, want: model.KnowledgeAnswerTypeOther},
		{name: "empty", arg: `""`, want: model.KnowledgeAnswerTypeOther},
		{name: "null", arg: `null`, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var action SearchAnswerAction
			err := json.Unmarshal([]byte(`{"questions":["q"],"type":`+tc.arg+`,"message":"m"}`), &action)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("the arguments were accepted with type %q, want a refusal", action.Type)
				}
				return
			}
			if err != nil {
				t.Fatalf("the model's arguments did not parse: %v", err)
			}

			got := model.KnowledgeAnswerType(action.Type)
			if got != tc.want {
				t.Errorf("stored and searched as %q, so a search for %q would miss it; wanted %q", got, tc.want, tc.want)
			}
			if !got.IsValid() {
				t.Errorf("%q is not a member of the enum, so no other door accepts it", got)
			}
		})
	}
}

// Reflects the schema as registry.go builds the live definition, so it is the one the model is shown.
func TestArgs_GraphitiSearchAction_SchemaCarriesEveryField(t *testing.T) {
	schema := reflector.Reflect(&GraphitiSearchAction{})

	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("failed to unmarshal schema into a generic map: %v", err)
	}

	properties, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no 'properties' object, got: %s", raw)
	}

	wantFields := []string{
		"search_type", "query", "max_results", "time_start", "time_end",
		"center_node_uuid", "max_depth", "node_labels", "edge_types",
		"diversity_level", "min_mentions", "recency_window", "message",
	}
	for _, field := range wantFields {
		if _, ok := properties[field]; !ok {
			t.Errorf("expected field %q in the JSON schema 'properties' but it is missing (full schema below)\n%s", field, raw)
		}
	}
	if len(properties) != len(wantFields) {
		t.Errorf("expected exactly %d properties, got %d: %v", len(wantFields), len(properties), slices.Collect(maps.Keys(properties)))
	}

	requiredRaw, ok := doc["required"].([]any)
	if !ok {
		t.Fatalf("schema has no 'required' array, got: %s", raw)
	}
	var required []string
	for _, r := range requiredRaw {
		s, ok := r.(string)
		if !ok {
			t.Fatalf("required entry is not a string: %v", r)
		}
		required = append(required, s)
	}
	wantRequired := []string{"search_type", "query", "message"}
	if len(required) != len(wantRequired) {
		t.Fatalf("expected required=%v, got %v", wantRequired, required)
	}
	for _, field := range wantRequired {
		if !slices.Contains(required, field) {
			t.Errorf("expected %q to be in the required list, got %v", field, required)
		}
	}

	centerNodeUUID, ok := properties["center_node_uuid"].(map[string]any)
	if !ok {
		t.Fatalf("center_node_uuid property is not an object: %v", properties["center_node_uuid"])
	}
	centerDesc, _ := centerNodeUUID["description"].(string)
	if !strings.Contains(centerDesc, "NEVER invent") {
		t.Errorf("expected center_node_uuid description to warn against inventing a value, got: %q", centerDesc)
	}
	if !strings.Contains(centerDesc, "UUID:") {
		t.Errorf("expected center_node_uuid description to reference the 'UUID:' field agents must copy from, got: %q", centerDesc)
	}

	nodeLabels, ok := properties["node_labels"].(map[string]any)
	if !ok {
		t.Fatalf("node_labels property is not an object: %v", properties["node_labels"])
	}
	nodeLabelsDesc, _ := nodeLabels["description"].(string)
	if !strings.Contains(nodeLabelsDesc, "PascalCase") {
		t.Errorf("expected node_labels description to mention PascalCase casing, got: %q", nodeLabelsDesc)
	}
}

func TestArgs_LenientLists_RepairAnUnderEscapedList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  string
		check func(t *testing.T, args string)
	}{
		{
			name: "a regex escaped for one nesting level instead of two",
			args: `{"subtasks":"[{\"title\":\"hunt\",\"description\":\"match FLAG\\{[a-f0-9]{64}\\}\"}]","message":"planned"}`,
			check: func(t *testing.T, args string) {
				var got SubtaskList
				if err := json.Unmarshal([]byte(args), &got); err != nil {
					t.Fatalf("the plan was refused: %v", err)
				}
				want := SubtaskInfos{{Title: "hunt", Description: `match FLAG\{[a-f0-9]{64}\}`}}
				if !reflect.DeepEqual(got.Subtasks, want) {
					t.Errorf("decoded %+v, want %+v", got.Subtasks, want)
				}
			},
		},
		{
			name: "hex escapes escaped for one nesting level instead of two",
			args: `{"operations":"[{\"op\":\"add\",\"title\":\"copy\",\"description\":\"bytes \\xff\\xfe\"}]","message":"one step"}`,
			check: func(t *testing.T, args string) {
				var got SubtaskPatch
				if err := json.Unmarshal([]byte(args), &got); err != nil {
					t.Fatalf("the batch was refused: %v", err)
				}
				want := SubtaskOperations{{Op: SubtaskOpAdd, Title: "copy", Description: `bytes \xff\xfe`}}
				if !reflect.DeepEqual(got.Operations, want) {
					t.Errorf("decoded %+v, want %+v", got.Operations, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.check(t, tt.args)
		})
	}
}

func TestArgs_RepairJSONStringEscapes_LeavesEverythingElseByteIdentical(t *testing.T) {
	t.Parallel()

	// A payload that already parses must survive untouched: the repair only ever
	// runs after a failure, and doubling inside a correct payload would corrupt it.
	unchanged := []string{
		`[{"title":"recon","description":"map the surface"}]`,
		`[{"description":"a quote \" and a newline \n and a tab \t"}]`,
		`[{"description":"a unicode escape \u00e9 stays one escape"}]`,
		`[{"description":"a slash \/ and a backslash \\ are both legal"}]`,
		`{"op":"add"}`,
		``,
	}
	for _, in := range unchanged {
		if got := repairJSONStringEscapes(in); got != in {
			t.Errorf("repair rewrote a legal payload\n in: %s\nout: %s", in, got)
		}
	}

	repaired := map[string]string{
		`["FLAG\{x\}"]`: `["FLAG\\{x\\}"]`,
		`["\xff"]`:      `["\\xff"]`,
		`["C:\users"]`:  `["C:\\users"]`,
		`["\u00zz"]`:    `["\\u00zz"]`,
		`["trailing\`:   `["trailing\\`,
	}
	for in, want := range repaired {
		if got := repairJSONStringEscapes(in); got != want {
			t.Errorf("repair of %s\n got: %s\nwant: %s", in, got, want)
		}
	}

	// A backslash outside a string literal is not the model's slip and is left alone.
	if got := repairJSONStringEscapes(`[\]`); got != `[\]` {
		t.Errorf("repair touched a backslash outside a string: %s", got)
	}
}

func TestArgs_TaskResult_RequiresAnExplicitSuccessFlag(t *testing.T) {
	t.Parallel()

	// An omitted flag must stay distinguishable from false: decoding it as false
	// records finished work as a failed task, and decoding it as true hides a
	// real failure.
	tests := []struct {
		name        string
		args        string
		wantValid   bool
		wantOutcome bool
	}{
		{name: "an explicit success", args: `{"success":true,"result":"r","message":"m"}`, wantValid: true, wantOutcome: true},
		{name: "an explicit failure", args: `{"success":false,"result":"r","message":"m"}`, wantValid: true, wantOutcome: false},
		{name: "a quoted success", args: `{"success":"true","result":"r","message":"m"}`, wantValid: true, wantOutcome: true},
		{name: "the flag omitted", args: `{"result":"SUCCESS - flag recovered","message":"m"}`, wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var task TaskResult
			if err := json.Unmarshal([]byte(tt.args), &task); err != nil {
				t.Fatalf("decoding failed: %v", err)
			}
			var done Done
			if err := json.Unmarshal([]byte(tt.args), &done); err != nil {
				t.Fatalf("decoding failed: %v", err)
			}

			taskErr, doneErr := task.Validate(), done.Validate()
			if tt.wantValid {
				if taskErr != nil {
					t.Fatalf("TaskResult was refused: %v", taskErr)
				}
				if doneErr != nil {
					t.Fatalf("Done was refused: %v", doneErr)
				}
				if got := task.Success.Bool(); got != tt.wantOutcome {
					t.Errorf("TaskResult success is %v, want %v", got, tt.wantOutcome)
				}
				if got := done.Success.Bool(); got != tt.wantOutcome {
					t.Errorf("Done success is %v, want %v", got, tt.wantOutcome)
				}

				return
			}

			for label, err := range map[string]error{"TaskResult": taskErr, "Done": doneErr} {
				if err == nil {
					t.Fatalf("%s accepted a report with no success flag", label)
				}
				if !strings.Contains(err.Error(), "success") {
					t.Errorf("%s refusal %q does not name the field", label, err)
				}
			}
			if task.Success != nil || done.Success != nil {
				t.Error("an omitted flag must decode to nil, not to a boolean")
			}
		})
	}
}
