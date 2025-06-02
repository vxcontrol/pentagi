package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Assistantlog is model to contain log record information from agents about their actions
// nolint:lll
type Assistantlog struct {
	ID           uint64             `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type         MsglogType         `form:"type" json:"type" validate:"valid,required" gorm:"type:MSGLOG_TYPE;NOT NULL"`
	Message      string             `form:"message" json:"message" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Thinking     string             `form:"thinking" json:"thinking" validate:"omitempty" gorm:"type:TEXT;NULL"`
	Result       string             `form:"result" json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	ResultFormat MsglogResultFormat `form:"result_format" json:"result_format" validate:"valid,required" gorm:"type:MSGLOG_RESULT_FORMAT;NOT NULL;default:plain"`
	FlowID       uint64             `form:"flow_id" json:"flow_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	AssistantID  uint64             `form:"assistant_id" json:"assistant_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt    time.Time          `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (al *Assistantlog) TableName() string {
	return "assistantlogs"
}

// Valid is function to control input/output data
func (al Assistantlog) Valid() error {
	return validate.Struct(al)
}

// Validate is function to use callback to control input/output data
func (al Assistantlog) Validate(db *gorm.DB) {
	if err := al.Valid(); err != nil {
		db.AddError(err)
	}
}
