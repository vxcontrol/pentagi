package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Screenshot is model to contain screenshot information
// nolint:lll
type Screenshot struct {
	ID        uint64    `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Name      string    `form:"name" json:"name" validate:"required" gorm:"type:TEXT;NOT NULL"`
	URL       string    `form:"url" json:"url" validate:"required" gorm:"type:TEXT;NOT NULL"`
	FlowID    uint64    `form:"flow_id" json:"flow_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL"`
	TaskID    *uint64   `form:"task_id,omitempty" json:"task_id,omitempty" validate:"omitempty,min=0,numeric" gorm:"type:BIGINT"`
	SubtaskID *uint64   `form:"subtask_id,omitempty" json:"subtask_id,omitempty" validate:"omitempty,min=0,numeric" gorm:"type:BIGINT"`
	CreatedAt time.Time `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (s *Screenshot) TableName() string {
	return "screenshots"
}

// Valid is function to control input/output data
func (s Screenshot) Valid() error {
	return validate.Struct(s)
}

// Validate is function to use callback to control input/output data
func (s Screenshot) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}
