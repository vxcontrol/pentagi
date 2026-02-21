package models

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jinzhu/gorm"
)

// TokenStatus represents the status of an API token
type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusRevoked TokenStatus = "revoked"
	TokenStatusExpired TokenStatus = "expired"
)

func (s TokenStatus) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s TokenStatus) Valid() error {
	switch s {
	case TokenStatusActive, TokenStatusRevoked, TokenStatusExpired:
		return nil
	default:
		return fmt.Errorf("invalid TokenStatus: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s TokenStatus) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// APIToken is model to contain API token metadata
// nolint:lll
type APIToken struct {
	ID        uint64      `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	TokenID   string      `form:"token_id" json:"token_id" validate:"required,len=10" gorm:"type:TEXT;NOT NULL;UNIQUE_INDEX"`
	UserID    uint64      `form:"user_id" json:"user_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	RoleID    uint64      `form:"role_id" json:"role_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	Name      *string     `form:"name,omitempty" json:"name,omitempty" validate:"omitempty,max=100" gorm:"type:TEXT"`
	TTL       uint64      `form:"ttl" json:"ttl" validate:"required,min=60,max=94608000" gorm:"type:BIGINT;NOT NULL"`
	Status    TokenStatus `form:"status" json:"status" validate:"valid,required" gorm:"type:TOKEN_STATUS;NOT NULL;default:'active'"`
	CreatedAt time.Time   `form:"created_at" json:"created_at" validate:"required" gorm:"type:TIMESTAMPTZ;NOT NULL;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time   `form:"updated_at" json:"updated_at" validate:"required" gorm:"type:TIMESTAMPTZ;NOT NULL;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time  `form:"deleted_at,omitempty" json:"deleted_at,omitempty" validate:"omitempty" sql:"index" gorm:"type:TIMESTAMPTZ"`
}

// TableName returns the table name string to guaranty use correct table
func (at *APIToken) TableName() string {
	return "api_tokens"
}

// Valid is function to control input/output data
func (at APIToken) Valid() error {
	if err := at.Status.Valid(); err != nil {
		return err
	}
	return validate.Struct(at)
}

// Validate is function to use callback to control input/output data
func (at APIToken) Validate(db *gorm.DB) {
	if err := at.Valid(); err != nil {
		db.AddError(err)
	}
}

// APITokenWithSecret is model to contain API token with the JWT token string (returned only on creation)
// nolint:lll
type APITokenWithSecret struct {
	APIToken `form:"" json:""`
	Token    string `form:"token" json:"token" validate:"required,jwt" gorm:"-"`
}

// Valid is function to control input/output data
func (ats APITokenWithSecret) Valid() error {
	if err := ats.APIToken.Valid(); err != nil {
		return err
	}
	return validate.Struct(ats)
}

// CreateAPITokenRequest is model to contain request data for creating an API token
// nolint:lll
type CreateAPITokenRequest struct {
	Name *string `form:"name,omitempty" json:"name,omitempty" validate:"omitempty,max=100"`
	TTL  uint64  `form:"ttl" json:"ttl" validate:"required,min=60,max=94608000"` // from 1 minute to 3 years
}

// Valid is function to control input/output data
func (catr CreateAPITokenRequest) Valid() error {
	return validate.Struct(catr)
}

// UpdateAPITokenRequest is model to contain request data for updating an API token
// nolint:lll
type UpdateAPITokenRequest struct {
	Name   *string     `form:"name,omitempty" json:"name,omitempty" validate:"omitempty,max=100"`
	Status TokenStatus `form:"status,omitempty" json:"status,omitempty" validate:"omitempty,valid"`
}

// Valid is function to control input/output data
func (uatr UpdateAPITokenRequest) Valid() error {
	if uatr.Status != "" {
		if err := uatr.Status.Valid(); err != nil {
			return err
		}
	}
	return validate.Struct(uatr)
}

// APITokenClaims is model to contain JWT claims for API tokens
// nolint:lll
type APITokenClaims struct {
	TokenID string `json:"tid" validate:"required,len=10"`
	RID     uint64 `json:"rid" validate:"min=0,max=10000"`
	UID     uint64 `json:"uid" validate:"min=0,max=10000"`
	UHASH   string `json:"uhash" validate:"required"`
	jwt.RegisteredClaims
}

// Valid is function to control input/output data
func (atc APITokenClaims) Valid() error {
	return validate.Struct(atc)
}
