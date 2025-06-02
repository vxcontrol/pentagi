package models

import "github.com/jinzhu/gorm"

// Prompt is model to contain prompt information
// nolint:lll
type Prompt struct {
	ID     uint64 `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type   string `form:"type" json:"type" validate:"required" gorm:"type:TEXT;NOT NULL"`
	UserID uint64 `form:"user_id" json:"user_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	Prompt string `form:"prompt" json:"prompt" validate:"required" gorm:"type:TEXT;NOT NULL"`
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
