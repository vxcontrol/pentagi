package tools

import (
	"encoding/json"
	"testing"
)

func boolPtr(b bool) *Bool {
	v := Bool(b)
	return &v
}

func int64Ptr(i int64) *Int64 {
	v := Int64(i)
	return &v
}

func TestBoolUnmarshalJSON(t *testing.T) {
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

func TestBoolMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    *Bool
		want string
	}{
		{name: "true value", b: boolPtr(true), want: "true"},
		{name: "false value", b: boolPtr(false), want: "false"},
		{name: "nil pointer", b: nil, want: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.b.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBoolBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    *Bool
		want bool
	}{
		{name: "true value", b: boolPtr(true), want: true},
		{name: "false value", b: boolPtr(false), want: false},
		{name: "nil pointer", b: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.b.Bool(); got != tt.want {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    *Bool
		want string
	}{
		{name: "true value", b: boolPtr(true), want: "true"},
		{name: "false value", b: boolPtr(false), want: "false"},
		{name: "nil pointer", b: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.b.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInt64UnmarshalJSON(t *testing.T) {
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
		{name: "empty string", input: `""`, wantErr: true},
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

func TestInt64MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    *Int64
		want string
	}{
		{name: "positive value", i: int64Ptr(42), want: "42"},
		{name: "negative value", i: int64Ptr(-7), want: "-7"},
		{name: "zero value", i: int64Ptr(0), want: "0"},
		{name: "nil pointer", i: nil, want: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.i.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestInt64Int(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    *Int64
		want int
	}{
		{name: "positive value", i: int64Ptr(42), want: 42},
		{name: "negative value", i: int64Ptr(-7), want: -7},
		{name: "zero value", i: int64Ptr(0), want: 0},
		{name: "nil pointer", i: nil, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.i.Int(); got != tt.want {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64Int64Method(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    *Int64
		want int64
	}{
		{name: "positive value", i: int64Ptr(42), want: 42},
		{name: "negative value", i: int64Ptr(-7), want: -7},
		{name: "zero value", i: int64Ptr(0), want: 0},
		{name: "nil pointer", i: nil, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.i.Int64(); got != tt.want {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64PtrInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		i       *Int64
		wantNil bool
		want    int64
	}{
		{name: "positive value", i: int64Ptr(42), wantNil: false, want: 42},
		{name: "negative value", i: int64Ptr(-7), wantNil: false, want: -7},
		{name: "zero value", i: int64Ptr(0), wantNil: false, want: 0},
		{name: "nil pointer", i: nil, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.i.PtrInt64()
			if tt.wantNil {
				if got != nil {
					t.Errorf("PtrInt64() = %v, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("PtrInt64() = nil, want non-nil")
			}
			if *got != tt.want {
				t.Errorf("PtrInt64() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func TestInt64String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    *Int64
		want string
	}{
		{name: "positive value", i: int64Ptr(42), want: "42"},
		{name: "negative value", i: int64Ptr(-7), want: "-7"},
		{name: "zero value", i: int64Ptr(0), want: "0"},
		{name: "nil pointer", i: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.i.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBoolJSONRoundTrip tests Bool marshal/unmarshal round-trip via struct embedding
func TestBoolJSONRoundTrip(t *testing.T) {
	t.Parallel()

	type container struct {
		Value Bool `json:"value"`
	}

	tests := []struct {
		name     string
		jsonData string
		want     Bool
	}{
		{name: "true from struct", jsonData: `{"value": true}`, want: true},
		{name: "false from struct", jsonData: `{"value": false}`, want: false},
		{name: "quoted true", jsonData: `{"value": "true"}`, want: true},
		{name: "quoted false", jsonData: `{"value": "false"}`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c container
			if err := json.Unmarshal([]byte(tt.jsonData), &c); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if c.Value != tt.want {
				t.Errorf("Value = %v, want %v", c.Value, tt.want)
			}

			data, err := json.Marshal(c)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			var c2 container
			if err := json.Unmarshal(data, &c2); err != nil {
				t.Fatalf("round-trip Unmarshal() error = %v", err)
			}
			if c2.Value != tt.want {
				t.Errorf("round-trip Value = %v, want %v", c2.Value, tt.want)
			}
		})
	}
}

// TestInt64JSONRoundTrip tests Int64 marshal/unmarshal round-trip via struct embedding
func TestInt64JSONRoundTrip(t *testing.T) {
	t.Parallel()

	type container struct {
		Value Int64 `json:"value"`
	}

	tests := []struct {
		name     string
		jsonData string
		want     Int64
	}{
		{name: "bare integer", jsonData: `{"value": 42}`, want: 42},
		{name: "negative integer", jsonData: `{"value": -99}`, want: -99},
		{name: "zero", jsonData: `{"value": 0}`, want: 0},
		{name: "quoted integer", jsonData: `{"value": "123"}`, want: 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c container
			if err := json.Unmarshal([]byte(tt.jsonData), &c); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if c.Value != tt.want {
				t.Errorf("Value = %v, want %v", c.Value, tt.want)
			}

			data, err := json.Marshal(c)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			var c2 container
			if err := json.Unmarshal(data, &c2); err != nil {
				t.Fatalf("round-trip Unmarshal() error = %v", err)
			}
			if c2.Value != tt.want {
				t.Errorf("round-trip Value = %v, want %v", c2.Value, tt.want)
			}
		})
	}
}

func TestSearchInMemoryAction_QuestionsUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "single question",
			json:    `{"questions": ["test query"], "message": "test"}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple questions",
			json:    `{"questions": ["query1", "query2", "query3"], "message": "test"}`,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "five questions max",
			json:    `{"questions": ["q1", "q2", "q3", "q4", "q5"], "message": "test"}`,
			wantLen: 5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var action SearchInMemoryAction
			err := json.Unmarshal([]byte(tt.json), &action)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(action.Questions) != tt.wantLen {
				t.Errorf("Questions length = %d, want %d", len(action.Questions), tt.wantLen)
			}
		})
	}
}

func TestSearchGuideAction_QuestionsUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "single question",
			json:    `{"questions": ["how to install tool"], "type": "install", "message": "test"}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple questions",
			json:    `{"questions": ["q1", "q2", "q3"], "type": "pentest", "message": "test"}`,
			wantLen: 3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var action SearchGuideAction
			err := json.Unmarshal([]byte(tt.json), &action)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(action.Questions) != tt.wantLen {
				t.Errorf("Questions length = %d, want %d", len(action.Questions), tt.wantLen)
			}
		})
	}
}

func TestSearchAnswerAction_QuestionsUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "single question",
			json:    `{"questions": ["what is exploit"], "type": "vulnerability", "message": "test"}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple questions",
			json:    `{"questions": ["q1", "q2"], "type": "tool", "message": "test"}`,
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var action SearchAnswerAction
			err := json.Unmarshal([]byte(tt.json), &action)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(action.Questions) != tt.wantLen {
				t.Errorf("Questions length = %d, want %d", len(action.Questions), tt.wantLen)
			}
		})
	}
}

func TestSearchCodeAction_QuestionsUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "single question",
			json:    `{"questions": ["python script for parsing"], "lang": "python", "message": "test"}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple questions",
			json:    `{"questions": ["bash script", "shell automation", "file processing"], "lang": "bash", "message": "test"}`,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "five questions",
			json:    `{"questions": ["q1", "q2", "q3", "q4", "q5"], "lang": "golang", "message": "test"}`,
			wantLen: 5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var action SearchCodeAction
			err := json.Unmarshal([]byte(tt.json), &action)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(action.Questions) != tt.wantLen {
				t.Errorf("Questions length = %d, want %d", len(action.Questions), tt.wantLen)
			}
		})
	}
}
