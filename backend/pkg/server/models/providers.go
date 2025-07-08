package models

import (
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/providers/provider"

	"github.com/jinzhu/gorm"
)

type ProviderType provider.ProviderType

func (s ProviderType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s ProviderType) Valid() error {
	providerType := provider.ProviderType(s)
	switch providerType {
	case provider.ProviderOpenAI,
		provider.ProviderAnthropic,
		provider.ProviderGemini,
		provider.ProviderBedrock,
		provider.ProviderOllama,
		provider.ProviderCustom:
		return nil
	default:
		return fmt.Errorf("invalid ProviderType: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s ProviderType) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Provider is model to contain provider configuration information
// nolint:lll
type Provider struct {
	ID        uint64          `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	UserID    uint64          `form:"user_id" json:"user_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	Type      ProviderType    `form:"type" json:"type" validate:"valid,required" gorm:"type:PROVIDER_TYPE;NOT NULL"`
	Name      string          `form:"name" json:"name" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Config    json.RawMessage `form:"config" json:"config" validate:"required" gorm:"type:JSON;NOT NULL"`
	CreatedAt time.Time       `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time       `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time      `form:"deleted_at,omitempty" json:"deleted_at,omitempty" validate:"omitempty" sql:"index" gorm:"type:TIMESTAMPTZ"`
}

// TableName returns the table name string to guaranty use correct table
func (p *Provider) TableName() string {
	return "providers"
}

// Valid is function to control input/output data
func (p Provider) Valid() error {
	return validate.Struct(p)
}

// Validate is function to use callback to control input/output data
func (p Provider) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// CreateProvider is model to contain provider creation payload
// nolint:lll
type CreateProvider struct {
	Config json.RawMessage `form:"config" json:"config" validate:"required" example:"{}"`
}

// Valid is function to control input/output data
func (cp CreateProvider) Valid() error {
	return validate.Struct(cp)
}

// PatchProvider is model to contain provider patching payload
// nolint:lll
type PatchProvider struct {
	Name   *string          `form:"name,omitempty" json:"name,omitempty" validate:"omitempty" example:"updated provider name"`
	Config *json.RawMessage `form:"config,omitempty" json:"config,omitempty" validate:"omitempty" example:"{}"`
}

// Valid is function to control input/output data
func (pp PatchProvider) Valid() error {
	return validate.Struct(pp)
}

// ProviderInfo is model to contain provider short information for display
// nolint:lll
type ProviderInfo struct {
	Name string       `form:"name" json:"name" validate:"required" example:"my openai provider"`
	Type ProviderType `form:"type" json:"type" validate:"valid,required" example:"openai"`
}

// Valid is function to control input/output data
func (p ProviderInfo) Valid() error {
	return validate.Struct(p)
}
