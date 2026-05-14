package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type ToolcallStatus string

const (
	ToolcallStatusReceived ToolcallStatus = "received"
	ToolcallStatusRunning  ToolcallStatus = "running"
	ToolcallStatusFinished ToolcallStatus = "finished"
	ToolcallStatusFailed   ToolcallStatus = "failed"
)

func (s ToolcallStatus) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s ToolcallStatus) Valid() error {
	switch s {
	case ToolcallStatusReceived,
		ToolcallStatusRunning,
		ToolcallStatusFinished,
		ToolcallStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid ToolcallStatus: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s ToolcallStatus) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Toolcall is model to contain tool call information
// nolint:lll
type Toolcall struct {
	ID              uint64         `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	CallID          string         `form:"call_id" json:"call_id" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Status          ToolcallStatus `form:"status" json:"status" validate:"valid,required" gorm:"type:TOOLCALL_STATUS;NOT NULL;default:'received'"`
	Name            string         `form:"name" json:"name" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Args            string         `form:"args" json:"args" validate:"required" gorm:"type:JSON;NOT NULL"`
	Result          string         `form:"result" json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	DurationSeconds float64        `form:"duration_seconds" json:"duration_seconds" validate:"min=0" gorm:"type:DOUBLE PRECISION;NOT NULL;default:0"`
	FlowID          uint64         `form:"flow_id" json:"flow_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL"`
	TaskID          *uint64        `form:"task_id,omitempty" json:"task_id,omitempty" validate:"omitnil,min=0" gorm:"type:BIGINT"`
	SubtaskID       *uint64        `form:"subtask_id,omitempty" json:"subtask_id,omitempty" validate:"omitnil,min=0" gorm:"type:BIGINT"`
	CreatedAt       time.Time      `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time      `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (tc *Toolcall) TableName() string {
	return "toolcalls"
}

// Valid is function to control input/output data
func (tc Toolcall) Valid() error {
	return validate.Struct(tc)
}

// Validate is function to use callback to control input/output data
func (tc Toolcall) Validate(db *gorm.DB) {
	if err := tc.Valid(); err != nil {
		db.AddError(err)
	}
}
