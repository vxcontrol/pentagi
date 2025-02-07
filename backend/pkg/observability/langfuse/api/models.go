// This file was auto-generated by Fern from our API Definition.

package api

import (
	json "encoding/json"
	fmt "fmt"
	internal "pentagi/pkg/observability/langfuse/api/internal"
	time "time"
)

type CreateModelRequest struct {
	// Name of the model definition. If multiple with the same name exist, they are applied in the following order: (1) custom over built-in, (2) newest according to startTime where model.startTime<observation.startTime
	ModelName string `json:"modelName" url:"-"`
	// Regex pattern which matches this model definition to generation.model. Useful in case of fine-tuned models. If you want to exact match, use `(?i)^modelname$`
	MatchPattern string `json:"matchPattern" url:"-"`
	// Apply only to generations which are newer than this ISO date.
	StartDate *time.Time `json:"startDate,omitempty" url:"-"`
	// Unit used by this model.
	Unit ModelUsageUnit `json:"unit" url:"-"`
	// Price (USD) per input unit
	InputPrice *float64 `json:"inputPrice,omitempty" url:"-"`
	// Price (USD) per output unit
	OutputPrice *float64 `json:"outputPrice,omitempty" url:"-"`
	// Price (USD) per total units. Cannot be set if input or output price is set.
	TotalPrice *float64 `json:"totalPrice,omitempty" url:"-"`
	// Optional. Tokenizer to be applied to observations which match to this model. See docs for more details.
	TokenizerId     *string     `json:"tokenizerId,omitempty" url:"-"`
	TokenizerConfig interface{} `json:"tokenizerConfig,omitempty" url:"-"`
}

func (c *CreateModelRequest) UnmarshalJSON(data []byte) error {
	type unmarshaler CreateModelRequest
	var body unmarshaler
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}
	*c = CreateModelRequest(body)
	return nil
}

func (c *CreateModelRequest) MarshalJSON() ([]byte, error) {
	type embed CreateModelRequest
	var marshaler = struct {
		embed
		StartDate *internal.DateTime `json:"startDate,omitempty"`
	}{
		embed:     embed(*c),
		StartDate: internal.NewOptionalDateTime(c.StartDate),
	}
	return json.Marshal(marshaler)
}

type ModelsListRequest struct {
	// page number, starts at 1
	Page *int `json:"-" url:"page,omitempty"`
	// limit of items per page
	Limit *int `json:"-" url:"limit,omitempty"`
}

// Model definition used for transforming usage into USD cost and/or tokenization.
type Model struct {
	Id string `json:"id" url:"id"`
	// Name of the model definition. If multiple with the same name exist, they are applied in the following order: (1) custom over built-in, (2) newest according to startTime where model.startTime<observation.startTime
	ModelName string `json:"modelName" url:"modelName"`
	// Regex pattern which matches this model definition to generation.model. Useful in case of fine-tuned models. If you want to exact match, use `(?i)^modelname$`
	MatchPattern string `json:"matchPattern" url:"matchPattern"`
	// Apply only to generations which are newer than this ISO date.
	StartDate *string `json:"startDate,omitempty" url:"startDate,omitempty"`
	// Unit used by this model.
	Unit ModelUsageUnit `json:"unit" url:"unit"`
	// Price (USD) per input unit
	InputPrice *float64 `json:"inputPrice,omitempty" url:"inputPrice,omitempty"`
	// Price (USD) per output unit
	OutputPrice *float64 `json:"outputPrice,omitempty" url:"outputPrice,omitempty"`
	// Price (USD) per total unit. Cannot be set if input or output price is set.
	TotalPrice *float64 `json:"totalPrice,omitempty" url:"totalPrice,omitempty"`
	// Optional. Tokenizer to be applied to observations which match to this model. See docs for more details.
	TokenizerId       *string     `json:"tokenizerId,omitempty" url:"tokenizerId,omitempty"`
	TokenizerConfig   interface{} `json:"tokenizerConfig,omitempty" url:"tokenizerConfig,omitempty"`
	IsLangfuseManaged bool        `json:"isLangfuseManaged" url:"isLangfuseManaged"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (m *Model) GetId() string {
	if m == nil {
		return ""
	}
	return m.Id
}

func (m *Model) GetModelName() string {
	if m == nil {
		return ""
	}
	return m.ModelName
}

func (m *Model) GetMatchPattern() string {
	if m == nil {
		return ""
	}
	return m.MatchPattern
}

func (m *Model) GetStartDate() *string {
	if m == nil {
		return nil
	}
	return m.StartDate
}

func (m *Model) GetUnit() ModelUsageUnit {
	if m == nil {
		return ""
	}
	return m.Unit
}

func (m *Model) GetInputPrice() *float64 {
	if m == nil {
		return nil
	}
	return m.InputPrice
}

func (m *Model) GetOutputPrice() *float64 {
	if m == nil {
		return nil
	}
	return m.OutputPrice
}

func (m *Model) GetTotalPrice() *float64 {
	if m == nil {
		return nil
	}
	return m.TotalPrice
}

func (m *Model) GetTokenizerId() *string {
	if m == nil {
		return nil
	}
	return m.TokenizerId
}

func (m *Model) GetTokenizerConfig() interface{} {
	if m == nil {
		return nil
	}
	return m.TokenizerConfig
}

func (m *Model) GetIsLangfuseManaged() bool {
	if m == nil {
		return false
	}
	return m.IsLangfuseManaged
}

func (m *Model) GetExtraProperties() map[string]interface{} {
	return m.extraProperties
}

func (m *Model) UnmarshalJSON(data []byte) error {
	type unmarshaler Model
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*m = Model(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *m)
	if err != nil {
		return err
	}
	m.extraProperties = extraProperties
	m.rawJSON = json.RawMessage(data)
	return nil
}

func (m *Model) String() string {
	if len(m.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(m.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(m); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", m)
}

type PaginatedModels struct {
	Data []*Model           `json:"data,omitempty" url:"data,omitempty"`
	Meta *UtilsMetaResponse `json:"meta,omitempty" url:"meta,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (p *PaginatedModels) GetData() []*Model {
	if p == nil {
		return nil
	}
	return p.Data
}

func (p *PaginatedModels) GetMeta() *UtilsMetaResponse {
	if p == nil {
		return nil
	}
	return p.Meta
}

func (p *PaginatedModels) GetExtraProperties() map[string]interface{} {
	return p.extraProperties
}

func (p *PaginatedModels) UnmarshalJSON(data []byte) error {
	type unmarshaler PaginatedModels
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = PaginatedModels(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *p)
	if err != nil {
		return err
	}
	p.extraProperties = extraProperties
	p.rawJSON = json.RawMessage(data)
	return nil
}

func (p *PaginatedModels) String() string {
	if len(p.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(p.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(p); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", p)
}
