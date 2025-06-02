package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type MsglogType string

const (
	MsglogTypeAnswer   MsglogType = "answer"
	MsglogTypeReport   MsglogType = "report"
	MsglogTypeThoughts MsglogType = "thoughts"
	MsglogTypeBrowser  MsglogType = "browser"
	MsglogTypeTerminal MsglogType = "terminal"
	MsglogTypeFile     MsglogType = "file"
	MsglogTypeSearch   MsglogType = "search"
	MsglogTypeAdvice   MsglogType = "advice"
	MsglogTypeAsk      MsglogType = "ask"
	MsglogTypeInput    MsglogType = "input"
	MsglogTypeDone     MsglogType = "done"
)

func (s MsglogType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s MsglogType) Valid() error {
	switch s {
	case MsglogTypeAnswer, MsglogTypeReport, MsglogTypeThoughts,
		MsglogTypeBrowser, MsglogTypeTerminal, MsglogTypeFile,
		MsglogTypeSearch, MsglogTypeAdvice, MsglogTypeAsk,
		MsglogTypeInput, MsglogTypeDone:
		return nil
	default:
		return fmt.Errorf("invalid MsglogType: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s MsglogType) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

type MsglogResultFormat string

const (
	MsglogResultFormatPlain    MsglogResultFormat = "plain"
	MsglogResultFormatMarkdown MsglogResultFormat = "markdown"
	MsglogResultFormatTerminal MsglogResultFormat = "terminal"
)

func (s MsglogResultFormat) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s MsglogResultFormat) Valid() error {
	switch s {
	case MsglogResultFormatPlain,
		MsglogResultFormatMarkdown,
		MsglogResultFormatTerminal:
		return nil
	default:
		return fmt.Errorf("invalid MsglogResultFormat: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s MsglogResultFormat) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Msglog is model to contain log record information from agents about their actions
// nolint:lll
type Msglog struct {
	ID           uint64             `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type         MsglogType         `form:"type" json:"type" validate:"valid,required" gorm:"type:MSGLOG_TYPE;NOT NULL"`
	Message      string             `form:"message" json:"message" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Thinking     string             `form:"thinking" json:"thinking" validate:"omitempty" gorm:"type:TEXT;NULL"`
	Result       string             `form:"result" json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	ResultFormat MsglogResultFormat `form:"result_format" json:"result_format" validate:"valid,required" gorm:"type:MSGLOG_RESULT_FORMAT;NOT NULL;default:plain"`
	FlowID       uint64             `form:"flow_id" json:"flow_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	TaskID       *uint64            `form:"task_id,omitempty" json:"task_id,omitempty" validate:"numeric,omitempty" gorm:"type:BIGINT;NOT NULL"`
	SubtaskID    *uint64            `form:"subtask_id,omitempty" json:"subtask_id,omitempty" validate:"numeric,omitempty" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt    time.Time          `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (ml *Msglog) TableName() string {
	return "msglogs"
}

// Valid is function to control input/output data
func (ml Msglog) Valid() error {
	return validate.Struct(ml)
}

// Validate is function to use callback to control input/output data
func (ml Msglog) Validate(db *gorm.DB) {
	if err := ml.Valid(); err != nil {
		db.AddError(err)
	}
}
