package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ProtoAuthToken is model to contain information to authorize vxproto connections
// nolint:lll
type ProtoAuthToken struct {
	Token       string    `form:"token" json:"token" validate:"jwt,required"`
	TTL         uint64    `form:"ttl" json:"ttl" validate:"min=1,max=94608000,required"`
	CreatedDate time.Time `form:"created_date,omitempty" json:"created_date,omitempty" validate:"omitempty"`
}

// Valid is function to control input/output data
func (pat ProtoAuthToken) Valid() error {
	return validate.Struct(pat)
}

// ProtoAuthToken is model to contain information to authorize vxproto connections
// nolint:lll
type ProtoAuthTokenRequest struct {
	TTL  uint64 `form:"ttl" json:"ttl" validate:"min=1,max=94608000,required" default:"31536000"`
	Type string `form:"type" json:"type" validate:"oneof=automation,required" default:"automation"`
}

// Valid is function to control input/output data
func (patr ProtoAuthTokenRequest) Valid() error {
	return validate.Struct(patr)
}

// ProtoAuthTokenClaims is model to contain token claims to authorize vxproto connections
// nolint:lll
type ProtoAuthTokenClaims struct {
	RID   uint64 `form:"rid" json:"rid" validate:"min=0,max=10000"`
	UID   uint64 `form:"uid" json:"uid" validate:"min=0,max=10000"`
	TID   string `form:"tid" json:"tid" validate:"required,oneof=local oauth"`
	UHASH string `form:"uhash" json:"uhash"`
	CPT   string `form:"cpt" json:"cpt" validate:"required,oneof=automation"`
	jwt.RegisteredClaims
}

// Valid is function to control input/output data
func (patc ProtoAuthTokenClaims) Valid() error {
	return validate.Struct(patc)
}
