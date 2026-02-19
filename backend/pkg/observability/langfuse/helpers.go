package langfuse

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/vxcontrol/langchaingo/llms"
)

const (
	firstVersion   = "v1"
	timeFormat8601 = "2006-01-02T15:04:05.000000Z"
)

var (
	ingestionCreateTrace      = "trace-create"
	ingestionCreateGeneration = "generation-create"
	ingestionUpdateGeneration = "generation-update"
	ingestionCreateSpan       = "span-create"
	ingestionUpdateSpan       = "span-update"
	ingestionCreateScore      = "score-create"
	ingestionCreateEvent      = "event-create"
	ingestionCreateAgent      = "agent-create"
	ingestionCreateTool       = "tool-create"
	ingestionCreateChain      = "chain-create"
	ingestionCreateEmbedding  = "embedding-create"
	ingestionCreateRetriever  = "retriever-create"
	ingestionCreateEvaluator  = "evaluator-create"
	ingestionCreateGuardrail  = "guardrail-create"
	ingestionPutLog           = "sdk-log"
)

type Metadata map[string]any

// mergeMaps combines two maps into a new map.
// Values from the second map (src) will override values from the first map (dst) for matching keys.
// Returns a new map with all combined key-value pairs.
// Handles nil values correctly: preserves original values without unnecessary allocations.
// If src is nil, returns dst as is (might be nil). If dst is nil but src is not, creates a copy of src.
func mergeMaps(dst, src map[string]any) map[string]any {
	if src == nil {
		return dst
	}

	if dst == nil {
		result := make(map[string]any, len(src))
		maps.Copy(result, src)
		return result
	}

	result := make(map[string]any, len(dst)+len(src))
	maps.Copy(result, dst)
	maps.Copy(result, src)
	return result
}

type ObservationLevel int

const (
	ObservationLevelDefault ObservationLevel = iota
	ObservationLevelDebug
	ObservationLevelWarning
	ObservationLevelError
)

func (e ObservationLevel) ToLangfuse() *api.ObservationLevel {
	var level api.ObservationLevel
	switch e {
	case ObservationLevelDebug:
		level = api.ObservationLevelDebug
	case ObservationLevelWarning:
		level = api.ObservationLevelWarning
	case ObservationLevelError:
		level = api.ObservationLevelError
	default:
		level = api.ObservationLevelDefault
	}
	return &level
}

type GenerationUsageUnit int

const (
	GenerationUsageUnitTokens GenerationUsageUnit = iota
	GenerationUsageUnitCharacters
	GenerationUsageUnitMilliseconds
	GenerationUsageUnitSeconds
	GenerationUsageUnitImages
	GenerationUsageUnitRequests
)

func (e GenerationUsageUnit) String() string {
	switch e {
	case GenerationUsageUnitTokens:
		return "TOKENS"
	case GenerationUsageUnitCharacters:
		return "CHARACTERS"
	case GenerationUsageUnitMilliseconds:
		return "MILLISECONDS"
	case GenerationUsageUnitSeconds:
		return "seconds"
	case GenerationUsageUnitImages:
		return "IMAGES"
	case GenerationUsageUnitRequests:
		return "REQUESTS"
	}
	return ""
}

func (e GenerationUsageUnit) ToLangfuse() *string {
	unit := e.String()
	if unit == "" {
		return nil
	}
	return &unit
}

type GenerationUsage struct {
	Input      int                 `json:"input,omitempty"`
	Output     int                 `json:"output,omitempty"`
	InputCost  *float64            `json:"input_cost,omitempty"`
	OutputCost *float64            `json:"output_cost,omitempty"`
	Unit       GenerationUsageUnit `json:"unit,omitempty"`
}

func (u *GenerationUsage) ToLangfuse() *api.IngestionUsage {
	if u == nil {
		return nil
	}

	var totalCost *float64
	if u.InputCost != nil {
		total := *u.InputCost
		totalCost = &total
	}
	if u.OutputCost != nil {
		total := *u.OutputCost
		if totalCost != nil {
			total += *totalCost
		}
		totalCost = &total
	}

	return &api.IngestionUsage{Usage: &api.Usage{
		Input:      u.Input,
		Output:     u.Output,
		Total:      u.Input + u.Output,
		InputCost:  u.InputCost,
		OutputCost: u.OutputCost,
		TotalCost:  totalCost,
		Unit:       u.Unit.ToLangfuse(),
	}}
}

type ModelParameters struct {
	// CandidateCount is the number of response candidates to generate.
	CandidateCount int `json:"candidate_count"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens"`
	// Temperature is the temperature for sampling, between 0 and 1.
	Temperature float64 `json:"temperature"`
	// StopWords is a list of words to stop on.
	StopWords []string `json:"stop_words"`
	// TopK is the number of tokens to consider for top-k sampling.
	TopK int `json:"top_k"`
	// TopP is the cumulative probability for top-p sampling.
	TopP float64 `json:"top_p"`
	// Seed is a seed for deterministic sampling.
	Seed int `json:"seed"`
	// MinLength is the minimum length of the generated text.
	MinLength int `json:"min_length"`
	// MaxLength is the maximum length of the generated text.
	MaxLength int `json:"max_length"`
	// N is how many chat completion choices to generate for each input message.
	N int `json:"n"`
	// RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty float64 `json:"repetition_penalty"`
	// FrequencyPenalty is the frequency penalty for sampling.
	FrequencyPenalty float64 `json:"frequency_penalty"`
	// PresencePenalty is the presence penalty for sampling.
	PresencePenalty float64 `json:"presence_penalty"`

	// JSONMode is a flag to enable JSON mode.
	JSONMode bool `json:"json"`
}

func (m *ModelParameters) ToLangfuse() map[string]*api.MapValue {
	if m == nil {
		return nil
	}

	parametersMap := make(map[string]any)
	parametersMap["temperature"] = fmt.Sprintf("%0.1f", m.Temperature)
	parametersMap["top_p"] = fmt.Sprintf("%0.1f", m.TopP)
	if m.CandidateCount != 0 {
		parametersMap["candidate_count"] = m.CandidateCount
	}
	if m.MaxTokens != 0 {
		parametersMap["max_tokens"] = m.MaxTokens
	} else {
		parametersMap["max_tokens"] = "inf"
	}
	if len(m.StopWords) > 0 {
		parametersMap["stop_words"] = m.StopWords
	}
	if m.TopK != 0 {
		parametersMap["top_k"] = m.TopK
	}
	if m.Seed != 0 {
		parametersMap["seed"] = m.Seed
	}
	if m.MinLength != 0 {
		parametersMap["min_length"] = m.MinLength
	}
	if m.MaxLength != 0 {
		parametersMap["max_length"] = m.MaxLength
	}
	if m.N != 0 {
		parametersMap["n"] = m.N
	}
	if m.RepetitionPenalty != 0 {
		parametersMap["repetition_penalty"] = fmt.Sprintf("%0.1f", m.RepetitionPenalty)
	}
	if m.FrequencyPenalty != 0 {
		parametersMap["frequency_penalty"] = fmt.Sprintf("%0.1f", m.FrequencyPenalty)
	}
	if m.PresencePenalty != 0 {
		parametersMap["presence_penalty"] = fmt.Sprintf("%0.1f", m.PresencePenalty)
	}
	if m.JSONMode {
		parametersMap["json"] = m.JSONMode
	}

	parametersData, err := json.Marshal(parametersMap)
	if err != nil {
		return nil
	}

	var parameters map[string]*api.MapValue
	if err := json.Unmarshal(parametersData, &parameters); err != nil {
		return nil
	}

	return parameters
}

func GetLangchainModelParameters(options []llms.CallOption) *ModelParameters {
	if len(options) == 0 {
		return nil
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	optsData, err := json.Marshal(opts)
	if err != nil {
		return nil
	}

	var parameters ModelParameters
	if err := json.Unmarshal(optsData, &parameters); err != nil {
		return nil
	}

	return &parameters
}

// newTraceID generates W3C Trace Context compliant trace ID
// Returns 32 lowercase hexadecimal characters
func newTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newSpanID generates W3C Trace Context compliant span/observation ID
// Returns 16 lowercase hexadecimal characters
func newSpanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getCurrentTime() time.Time {
	return time.Now().UTC()
}

func getCurrentTimeString() string {
	return getCurrentTime().Format(timeFormat8601)
}

func getCurrentTimeRef() *time.Time {
	return getTimeRef(getCurrentTime())
}

func getTimeRef(time time.Time) *time.Time {
	return &time
}

func getTimeRefString(time *time.Time) string {
	if time == nil {
		return getCurrentTimeString()
	}
	return time.Format(timeFormat8601)
}

func getStringRef(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getIntRef(i int) *int {
	return &i
}

func getBoolRef(b bool) *bool {
	return &b
}
