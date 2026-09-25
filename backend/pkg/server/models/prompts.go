package models

import (
	"fmt"
	"time"

	"pentagi/pkg/templates"

	"github.com/jinzhu/gorm"
)

// PromptType is an alias for templates.PromptType with validation methods for GORM
type PromptType templates.PromptType

// String returns the string representation of PromptType
func (s PromptType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s PromptType) Valid() error {
	if _, ok := templates.PromptVariables[templates.PromptType(s)]; !ok {
		return fmt.Errorf("invalid PromptType: %s", s)
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (s PromptType) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Prompt is model to contain prompt information
// nolint:lll
type Prompt struct {
	ID        uint64     `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type      PromptType `form:"type" json:"type" validate:"valid,required" gorm:"type:PROMPT_TYPE;NOT NULL"`
	UserID    uint64     `form:"user_id" json:"user_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	Prompt    string     `form:"prompt" json:"prompt" validate:"required" gorm:"type:TEXT;NOT NULL"`
	CreatedAt time.Time  `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (p *Prompt) TableName() string {
	return "prompts"
}

// Valid is function to control input/output data
func (p Prompt) Valid() error {
	return validate.Struct(p)
}

// Validate is function to use callback to control input/output data
func (p Prompt) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// PatchPrompt is model to contain prompt patching paylaod
type PatchPrompt struct {
	Prompt string `form:"prompt" json:"prompt" validate:"required"`
}

// Valid is function to control input/output data
func (pp PatchPrompt) Valid() error {
	return validate.Struct(pp)
}
