package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type SubtaskStatus string

const (
	SubtaskStatusCreated  SubtaskStatus = "created"
	SubtaskStatusRunning  SubtaskStatus = "running"
	SubtaskStatusWaiting  SubtaskStatus = "waiting"
	SubtaskStatusFinished SubtaskStatus = "finished"
	SubtaskStatusFailed   SubtaskStatus = "failed"
)

func (s SubtaskStatus) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s SubtaskStatus) Valid() error {
	switch s {
	case SubtaskStatusCreated,
		SubtaskStatusRunning,
		SubtaskStatusWaiting,
		SubtaskStatusFinished,
		SubtaskStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid SubtaskStatus: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s SubtaskStatus) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Subtask is model to contain subtask information
// nolint:lll
type Subtask struct {
	ID          uint64        `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Status      SubtaskStatus `form:"status" json:"status" validate:"valid,required" gorm:"type:SUBTASK_STATUS;NOT NULL;default:'created'"`
	Title       string        `form:"title" json:"title" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Description string        `form:"description" json:"description" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Context     string        `form:"context" json:"context" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	Result      string        `form:"result" json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	TaskID      uint64        `form:"task_id" json:"task_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt   time.Time     `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time     `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (s *Subtask) TableName() string {
	return "subtasks"
}

// Valid is function to control input/output data
func (s Subtask) Valid() error {
	return validate.Struct(s)
}

// Validate is function to use callback to control input/output data
func (s Subtask) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}
