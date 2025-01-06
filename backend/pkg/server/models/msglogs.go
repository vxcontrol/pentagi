package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type MsglogType string

const (
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

type MsglogResultFormat string

const (
	MsglogResultFormatPlain    MsglogResultFormat = "plain"
	MsglogResultFormatMarkdown MsglogResultFormat = "markdown"
	MsglogResultFormatTerminal MsglogResultFormat = "terminal"
)

func (s MsglogResultFormat) String() string {
	return string(s)
}

// Msglog is model to contain log record information from agents about their actions
// nolint:lll
type Msglog struct {
	ID           uint64             `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type         MsglogType         `form:"type" json:"type" validate:"oneof=thoughts browser terminal file search advice ask input done,required" gorm:"type:MSGLOG_TYPE;NOT NULL"`
	Message      string             `form:"message" json:"message" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Result       string             `form:"result" json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	ResultFormat MsglogResultFormat `form:"result_format" json:"result_format" validate:"required" gorm:"type:MSGLOG_RESULT_FORMAT;NOT NULL;default:plain"`
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
