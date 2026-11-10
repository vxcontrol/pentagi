package langfuse

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/vxcontrol/langchaingo/llms"
)

const (
	firstVersion   = "v1"
	timeFormat8601 = "2006-01-02T15:04:05.000000Z"

	otelObservationTypeSpan       = "span"
	otelObservationTypeGeneration = "generation"
	otelObservationTypeEvent      = "event"
	otelObservationTypeEmbedding  = "embedding"
	otelObservationTypeAgent      = "agent"
	otelObservationTypeTool       = "tool"
	otelObservationTypeChain      = "chain"
	otelObservationTypeRetriever  = "retriever"
	otelObservationTypeGuardrail  = "guardrail"
	otelObservationTypeEvaluator  = "evaluator"

	otelAttrObservationType            = "langfuse.observation.type"
	otelAttrObservationLevel           = "langfuse.observation.level"
	otelAttrObservationStatusMessage   = "langfuse.observation.status_message"
	otelAttrObservationInput           = "langfuse.observation.input"
	otelAttrObservationOutput          = "langfuse.observation.output"
	otelAttrObservationMetadataPrefix  = "langfuse.observation.metadata."
	otelAttrObservationModelName       = "langfuse.observation.model.name"
	otelAttrObservationModelParameters = "langfuse.observation.model.parameters"
	otelAttrObservationUsageDetails    = "langfuse.observation.usage_details"
	otelAttrObservationCostDetails     = "langfuse.observation.cost_details"
	otelAttrObservationPromptName      = "langfuse.observation.prompt.name"
	otelAttrObservationPromptVersion   = "langfuse.observation.prompt.version"
	otelAttrObservationCompletionStart = "langfuse.observation.completion_start_time"

	otelAttrTraceName           = "langfuse.trace.name"
	otelAttrTracePublic         = "langfuse.trace.public"
	otelAttrTraceTags           = "langfuse.trace.tags"
	otelAttrTraceMetadataPrefix = "langfuse.trace.metadata."
	otelAttrUserID              = "langfuse.user.id"
	otelAttrSessionID           = "langfuse.session.id"
	otelAttrRelease             = "langfuse.release"
	otelAttrVersion             = "langfuse.version"
	otelAttrEnvironment         = "langfuse.environment"
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
		return "SECONDS"
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
	CandidateCount *int `json:"candidate_count,omitempty"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Temperature is the temperature for sampling, between 0 and 1.
	Temperature *float64 `json:"temperature,omitempty"`
	// StopWords is a list of words to stop on.
	StopWords []string `json:"stop_words"`
	// TopK is the number of tokens to consider for top-k sampling.
	TopK *int `json:"top_k,omitempty"`
	// TopP is the cumulative probability for top-p sampling.
	TopP *float64 `json:"top_p,omitempty"`
	// MinP is the minimum probability for top-p sampling.
	MinP *float64 `json:"min_p,omitempty"`
	// Seed is a seed for deterministic sampling.
	Seed *int `json:"seed,omitempty"`
	// MinLength is the minimum length of the generated text.
	MinLength *int `json:"min_length,omitempty"`
	// MaxLength is the maximum length of the generated text.
	MaxLength *int `json:"max_length,omitempty"`
	// N is how many chat completion choices to generate for each input message.
	N *int `json:"n,omitempty"`
	// RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`
	// FrequencyPenalty is the frequency penalty for sampling.
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty is the presence penalty for sampling.
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// JSONMode is a flag to enable JSON mode.
	JSONMode bool `json:"json"`
}

func (m *ModelParameters) ToLangfuse() map[string]*api.MapValue {
	if m == nil {
		return nil
	}

	parametersMap := make(map[string]any)
	if m.Temperature != nil {
		parametersMap["temperature"] = fmt.Sprintf("%0.1f", *m.Temperature)
	}
	if m.TopP != nil {
		parametersMap["top_p"] = fmt.Sprintf("%0.1f", *m.TopP)
	}
	if m.MinP != nil {
		parametersMap["min_p"] = fmt.Sprintf("%0.1f", *m.MinP)
	}
	if m.CandidateCount != nil {
		parametersMap["candidate_count"] = *m.CandidateCount
	}
	if m.MaxTokens != nil {
		parametersMap["max_tokens"] = *m.MaxTokens
	} else {
		parametersMap["max_tokens"] = "inf"
	}
	if len(m.StopWords) > 0 {
		parametersMap["stop_words"] = m.StopWords
	}
	if m.TopK != nil {
		parametersMap["top_k"] = *m.TopK
	}
	if m.Seed != nil {
		parametersMap["seed"] = *m.Seed
	}
	if m.MinLength != nil {
		parametersMap["min_length"] = *m.MinLength
	}
	if m.MaxLength != nil {
		parametersMap["max_length"] = *m.MaxLength
	}
	if m.N != nil {
		parametersMap["n"] = *m.N
	}
	if m.RepetitionPenalty != nil {
		parametersMap["repetition_penalty"] = fmt.Sprintf("%0.1f", *m.RepetitionPenalty)
	}
	if m.FrequencyPenalty != nil {
		parametersMap["frequency_penalty"] = fmt.Sprintf("%0.1f", *m.FrequencyPenalty)
	}
	if m.PresencePenalty != nil {
		parametersMap["presence_penalty"] = fmt.Sprintf("%0.1f", *m.PresencePenalty)
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

type otelSpanParams struct {
	TraceID             string
	ObservationID       string
	ParentObservationID string
	Name                string
	ObservationType     string
	StartTime           *time.Time
	EndTime             *time.Time
	CompletionStartTime *time.Time
	Level               ObservationLevel
	Status              *string
	Version             *string
	Input               any
	Output              any
	Metadata            Metadata
	Model               *string
	ModelParameters     *ModelParameters
	Usage               *GenerationUsage
	PromptName          *string
	PromptVersion       *int
	TraceCtx            *TraceContext
}

func otelStringAttr(key, val string) *api.OtelAttribute {
	k := key
	v := val
	return &api.OtelAttribute{
		Key:   &k,
		Value: &api.OtelAttributeValue{StringValue: &v},
	}
}

func otelIntAttr(key string, val int) *api.OtelAttribute {
	k := key
	v := val
	return &api.OtelAttribute{
		Key:   &k,
		Value: &api.OtelAttributeValue{IntValue: &v},
	}
}

func otelFloatAttr(key string, val float64) *api.OtelAttribute {
	k := key
	v := val
	return &api.OtelAttribute{
		Key:   &k,
		Value: &api.OtelAttributeValue{DoubleValue: &v},
	}
}

func otelBoolAttr(key string, val bool) *api.OtelAttribute {
	k := key
	v := val
	return &api.OtelAttribute{
		Key:   &k,
		Value: &api.OtelAttributeValue{BoolValue: &v},
	}
}

func otelJSONOrStringAttr(key string, val any) *api.OtelAttribute {
	if val == nil {
		return nil
	}
	if s, ok := val.(string); ok {
		return otelStringAttr(key, s)
	}
	data, err := json.Marshal(val)
	if err != nil {
		return otelStringAttr(key, fmt.Sprintf("%v", val))
	}
	return otelStringAttr(key, string(data))
}

func otelAnyAttr(key string, val any) *api.OtelAttribute {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		return otelStringAttr(key, v)
	case bool:
		return otelBoolAttr(key, v)
	case int:
		return otelIntAttr(key, v)
	case int64:
		return otelIntAttr(key, int(v))
	case int32:
		return otelIntAttr(key, int(v))
	case float64:
		return otelFloatAttr(key, v)
	case float32:
		return otelFloatAttr(key, float64(v))
	default:
		return otelJSONOrStringAttr(key, v)
	}
}

func buildOtelSpan(p otelSpanParams) *api.OtelSpan {
	startTime := getCurrentTime()
	if p.StartTime != nil {
		startTime = p.StartTime.UTC()
	}
	endTime := startTime
	if p.EndTime != nil {
		endTime = p.EndTime.UTC()
	}

	var attrs []*api.OtelAttribute
	appendAttr := func(attr *api.OtelAttribute) {
		if attr != nil {
			attrs = append(attrs, attr)
		}
	}

	obsType := p.ObservationType
	if obsType == "" {
		obsType = otelObservationTypeSpan
	}
	appendAttr(otelStringAttr(otelAttrObservationType, obsType))

	levelStr := string(*p.Level.ToLangfuse())
	appendAttr(otelStringAttr(otelAttrObservationLevel, levelStr))

	if p.Status != nil {
		appendAttr(otelStringAttr(otelAttrObservationStatusMessage, *p.Status))
	}

	inputVal := p.Input
	if inputVal == nil && p.ParentObservationID == "" && p.TraceCtx != nil && p.TraceCtx.Input != nil {
		inputVal = p.TraceCtx.Input
	}
	appendAttr(otelJSONOrStringAttr(otelAttrObservationInput, inputVal))

	outputVal := p.Output
	if outputVal == nil && p.ParentObservationID == "" && p.TraceCtx != nil && p.TraceCtx.Output != nil {
		outputVal = p.TraceCtx.Output
	}
	appendAttr(otelJSONOrStringAttr(otelAttrObservationOutput, outputVal))

	if len(p.Metadata) > 0 {
		for _, k := range slices.Sorted(maps.Keys(p.Metadata)) {
			appendAttr(otelAnyAttr(otelAttrObservationMetadataPrefix+k, p.Metadata[k]))
		}
	}

	if p.Model != nil && *p.Model != "" {
		appendAttr(otelStringAttr(otelAttrObservationModelName, *p.Model))
	}

	if params := p.ModelParameters.ToLangfuse(); params != nil {
		appendAttr(otelJSONOrStringAttr(otelAttrObservationModelParameters, params))
	}

	if usage := p.Usage.ToLangfuse(); usage != nil && usage.Usage != nil {
		u := usage.Usage
		usageDetails := map[string]int{
			"input":  u.Input,
			"output": u.Output,
			"total":  u.Total,
		}
		appendAttr(otelJSONOrStringAttr(otelAttrObservationUsageDetails, usageDetails))

		if u.InputCost != nil || u.OutputCost != nil || u.TotalCost != nil {
			costDetails := make(map[string]float64)
			if u.InputCost != nil {
				costDetails["input"] = *u.InputCost
			}
			if u.OutputCost != nil {
				costDetails["output"] = *u.OutputCost
			}
			if u.TotalCost != nil {
				costDetails["total"] = *u.TotalCost
			}
			appendAttr(otelJSONOrStringAttr(otelAttrObservationCostDetails, costDetails))
		}
	}

	if p.PromptName != nil && *p.PromptName != "" {
		appendAttr(otelStringAttr(otelAttrObservationPromptName, *p.PromptName))
	}
	if p.PromptVersion != nil {
		appendAttr(otelIntAttr(otelAttrObservationPromptVersion, *p.PromptVersion))
	}
	if p.CompletionStartTime != nil {
		appendAttr(otelStringAttr(otelAttrObservationCompletionStart, getTimeRefString(p.CompletionStartTime)))
	}

	version := p.Version
	if (version == nil || *version == "") && p.TraceCtx != nil && p.TraceCtx.Version != nil {
		version = p.TraceCtx.Version
	}
	if version != nil && *version != "" {
		appendAttr(otelStringAttr(otelAttrVersion, *version))
	}

	if p.TraceCtx != nil {
		if p.TraceCtx.Name != nil && *p.TraceCtx.Name != "" {
			appendAttr(otelStringAttr(otelAttrTraceName, *p.TraceCtx.Name))
		}
		if p.TraceCtx.UserID != nil && *p.TraceCtx.UserID != "" {
			appendAttr(otelStringAttr(otelAttrUserID, *p.TraceCtx.UserID))
		}
		if p.TraceCtx.SessionID != nil && *p.TraceCtx.SessionID != "" {
			appendAttr(otelStringAttr(otelAttrSessionID, *p.TraceCtx.SessionID))
		}
		if p.TraceCtx.Public != nil {
			appendAttr(otelBoolAttr(otelAttrTracePublic, *p.TraceCtx.Public))
		}
		if len(p.TraceCtx.Tags) > 0 {
			appendAttr(otelJSONOrStringAttr(otelAttrTraceTags, p.TraceCtx.Tags))
		}
		if len(p.TraceCtx.Metadata) > 0 {
			for _, k := range slices.Sorted(maps.Keys(p.TraceCtx.Metadata)) {
				appendAttr(otelAnyAttr(otelAttrTraceMetadataPrefix+k, p.TraceCtx.Metadata[k]))
			}
		}
	}

	span := &api.OtelSpan{
		TraceID:           p.TraceID,
		SpanID:            p.ObservationID,
		Name:              getStringRef(p.Name),
		Kind:              getIntRef(1),
		StartTimeUnixNano: strconv.FormatInt(startTime.UnixNano(), 10),
		EndTimeUnixNano:   strconv.FormatInt(endTime.UnixNano(), 10),
		Attributes:        attrs,
	}
	if p.ParentObservationID != "" {
		span.ParentSpanID = p.ParentObservationID
	}
	if p.Level == ObservationLevelError {
		status := map[string]any{"code": 2}
		if p.Status != nil {
			status["message"] = *p.Status
		}
		span.Status = status
	} else if p.Status != nil {
		span.Status = map[string]any{"code": 1, "message": *p.Status}
	}

	return span
}
