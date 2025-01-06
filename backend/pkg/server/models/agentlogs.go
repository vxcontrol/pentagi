package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Agentlog is model to contain agent task and result information
// nolint:lll
type Agentlog struct {
	ID        uint64       `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Initiator MsgchainType `json:"initiator" validate:"valid,required" gorm:"type:MSGCHAIN_TYPE;NOT NULL"`
	Executor  MsgchainType `json:"executor" validate:"valid,required" gorm:"type:MSGCHAIN_TYPE;NOT NULL"`
	Task      string       `json:"task" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Result    string       `json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	FlowID    uint64       `form:"flow_id" json:"flow_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	TaskID    *uint64      `form:"task_id,omitempty" json:"task_id,omitempty" validate:"numeric,omitempty" gorm:"type:BIGINT;NOT NULL"`
	SubtaskID *uint64      `form:"subtask_id,omitempty" json:"subtask_id,omitempty" validate:"numeric,omitempty" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt time.Time    `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (ml *Agentlog) TableName() string {
	return "agentlogs"
}

// Valid is function to control input/output data
func (ml Agentlog) Valid() error {
	return validate.Struct(ml)
}

// Validate is function to use callback to control input/output data
func (ml Agentlog) Validate(db *gorm.DB) {
	if err := ml.Valid(); err != nil {
		db.AddError(err)
	}
}
