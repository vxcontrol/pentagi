package models

import (
	"fmt"
	"pentagi/pkg/tools"
	"time"

	"github.com/jinzhu/gorm"
)

type AssistantStatus string

const (
	AssistantStatusCreated  AssistantStatus = "created"
	AssistantStatusRunning  AssistantStatus = "running"
	AssistantStatusWaiting  AssistantStatus = "waiting"
	AssistantStatusFinished AssistantStatus = "finished"
	AssistantStatusFailed   AssistantStatus = "failed"
)

func (s AssistantStatus) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s AssistantStatus) Valid() error {
	switch s {
	case AssistantStatusCreated,
		AssistantStatusRunning,
		AssistantStatusWaiting,
		AssistantStatusFinished,
		AssistantStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid AssistantStatus: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s AssistantStatus) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Assistant is model to contain assistant information
// nolint:lll
type Assistant struct {
	ID                 uint64           `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Status             AssistantStatus  `form:"status" json:"status" validate:"valid,required" gorm:"type:ASSISTANT_STATUS;NOT NULL;default:'created'"`
	Title              string           `form:"title" json:"title" validate:"required" gorm:"type:TEXT;NOT NULL;default:'untitled'"`
	Model              string           `form:"model" json:"model" validate:"max=70,required" gorm:"type:TEXT;NOT NULL"`
	ModelProviderName  string           `form:"model_provider_name" json:"model_provider_name" validate:"max=70,required" gorm:"type:TEXT;NOT NULL"`
	ModelProviderType  ProviderType     `form:"model_provider_type" json:"model_provider_type" validate:"valid,required" gorm:"type:PROVIDER_TYPE;NOT NULL"`
	Language           string           `form:"language" json:"language" validate:"max=70,required" gorm:"type:TEXT;NOT NULL"`
	Functions          *tools.Functions `form:"functions,omitempty" json:"functions,omitempty" validate:"omitempty,valid" gorm:"type:JSON;NOT NULL;default:'{}'"`
	FlowID             uint64           `form:"flow_id" json:"flow_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL"`
	MsgchainID         *uint64          `form:"msgchain_id" json:"msgchain_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	TraceID            *string          `form:"trace_id" json:"trace_id" validate:"max=70,required" gorm:"type:TEXT;NOT NULL"`
	ToolCallIDTemplate string           `form:"tool_call_id_template" json:"tool_call_id_template" validate:"max=70,required" gorm:"type:TEXT;NOT NULL"`
	UseAgents          bool             `form:"use_agents" json:"use_agents" validate:"omitempty" gorm:"type:BOOLEAN;NOT NULL;default:false"`
	CreatedAt          time.Time        `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt          time.Time        `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	DeletedAt          *time.Time       `form:"deleted_at,omitempty" json:"deleted_at,omitempty" validate:"omitempty" sql:"index" gorm:"type:TIMESTAMPTZ"`
}

// TableName returns the table name string to guaranty use correct table
func (a *Assistant) TableName() string {
	return "assistants"
}

// Valid is function to control input/output data
func (a Assistant) Valid() error {
	return validate.Struct(a)
}

// Validate is function to use callback to control input/output data
func (a Assistant) Validate(db *gorm.DB) {
	if err := a.Valid(); err != nil {
		db.AddError(err)
	}
}

// CreateAssistant is model to contain assistant creation paylaod
// nolint:lll
type CreateAssistant struct {
	Input     string           `form:"input" json:"input" validate:"required" example:"user input for running assistant"`
	Provider  string           `form:"provider" json:"provider" validate:"required" example:"openai"`
	UseAgents bool             `form:"use_agents" json:"use_agents" validate:"omitempty" example:"true"`
	Functions *tools.Functions `form:"functions,omitempty" json:"functions,omitempty" validate:"omitempty,valid"`
}

// Valid is function to control input/output data
func (ca CreateAssistant) Valid() error {
	return validate.Struct(ca)
}

// PatchAssistant is model to contain assistant patching paylaod
// nolint:lll
type PatchAssistant struct {
	Action    string  `form:"action" json:"action" validate:"required,oneof=stop input" enums:"stop,input" default:"stop"`
	Input     *string `form:"input,omitempty" json:"input,omitempty" validate:"required_if=Action input" example:"user input for waiting assistant"`
	UseAgents bool    `form:"use_agents" json:"use_agents" validate:"omitempty" example:"true"`
}

// Valid is function to control input/output data
func (pa PatchAssistant) Valid() error {
	return validate.Struct(pa)
}

// AssistantFlow is model to contain assistant information linked with flow
// nolint:lll
type AssistantFlow struct {
	Flow      Flow `form:"flow,omitempty" json:"flow,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	Assistant `form:"" json:""`
}

// Valid is function to control input/output data
func (af AssistantFlow) Valid() error {
	if err := af.Flow.Valid(); err != nil {
		return err
	}
	return af.Assistant.Valid()
}

// Validate is function to use callback to control input/output data
func (af AssistantFlow) Validate(db *gorm.DB) {
	if err := af.Valid(); err != nil {
		db.AddError(err)
	}
}
