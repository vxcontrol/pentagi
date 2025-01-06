package langfuse

import (
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
)

const (
	firstVersion   = "v1"
	timeFormat8601 = "2006-01-02T15:04:05.000000Z"
)

var (
	ingestionCreateTrace       = getStringRef("trace-create")
	ingestionCreateScore       = getStringRef("score-create")
	ingestionPutLog            = getStringRef("sdk-log")
	ingestionCreateObservation = getStringRef("observation-create")
	ingestionUpdateObservation = getStringRef("observation-update")
	ingestionCreateGeneration  = getStringRef("generation-create")
	ingestionUpdateGeneration  = getStringRef("generation-update")
)

type Metadata map[string]any

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
		return "tokens"
	case GenerationUsageUnitCharacters:
		return "characters"
	case GenerationUsageUnitMilliseconds:
		return "milliseconds"
	case GenerationUsageUnitSeconds:
		return "seconds"
	case GenerationUsageUnitImages:
		return "images"
	case GenerationUsageUnitRequests:
		return "requests"
	}
	return ""
}

func (e GenerationUsageUnit) ToLangfuse() *api.ModelUsageUnit {
	var unit api.ModelUsageUnit
	switch e {
	case GenerationUsageUnitTokens:
		unit = api.ModelUsageUnitTokens
	case GenerationUsageUnitCharacters:
		unit = api.ModelUsageUnitCharacters
	case GenerationUsageUnitMilliseconds:
		unit = api.ModelUsageUnitMilliseconds
	case GenerationUsageUnitSeconds:
		unit = api.ModelUsageUnitSeconds
	case GenerationUsageUnitImages:
		unit = api.ModelUsageUnitImages
	case GenerationUsageUnitRequests:
		unit = api.ModelUsageUnitRequests
	default:
		return nil
	}
	return &unit
}

type GenerationUsage struct {
	Input  int                 `json:"input,omitempty"`
	Output int                 `json:"output,omitempty"`
	Unit   GenerationUsageUnit `json:"unit,omitempty"`
}

func (u *GenerationUsage) ToLangfuse() *api.IngestionUsage {
	if u == nil {
		return nil
	}

	return api.NewIngestionUsageFromUsage(&api.Usage{
		Input:  getIntRef(u.Input),
		Output: getIntRef(u.Output),
		Total:  getIntRef(u.Input + u.Output),
		Unit:   u.Unit.ToLangfuse(),
	})
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

	parametersMap := make(map[string]interface{})
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

func newID() string {
	return uuid.New().String()
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
	return &s
}

func getIntRef(i int) *int {
	return &i
}

func getBoolRef(b bool) *bool {
	return &b
}
