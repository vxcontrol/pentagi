package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type VecstoreActionType string

const (
	VecstoreActionTypeRetrieve VecstoreActionType = "retrieve"
	VecstoreActionTypeStore    VecstoreActionType = "store"
)

func (s VecstoreActionType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s VecstoreActionType) Valid() error {
	switch s {
	case VecstoreActionTypeRetrieve, VecstoreActionTypeStore:
		return nil
	default:
		return fmt.Errorf("invalid VecstoreActionType: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s VecstoreActionType) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Vecstorelog is model to contain vecstore action information
// nolint:lll
type Vecstorelog struct {
	ID        uint64             `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Initiator MsgchainType       `json:"initiator" validate:"valid,required" gorm:"type:MSGCHAIN_TYPE;NOT NULL"`
	Executor  MsgchainType       `json:"executor" validate:"valid,required" gorm:"type:MSGCHAIN_TYPE;NOT NULL"`
	Filter    string             `json:"filter" validate:"required" gorm:"type:JSON;NOT NULL"`
	Query     string             `json:"query" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Action    VecstoreActionType `json:"action" validate:"valid,required" gorm:"type:VECSTORE_ACTION_TYPE;NOT NULL"`
	Result    string             `json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL"`
	FlowID    uint64             `form:"flow_id" json:"flow_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL"`
	TaskID    *uint64            `form:"task_id,omitempty" json:"task_id,omitempty" validate:"omitnil,min=0" gorm:"type:BIGINT;NOT NULL"`
	SubtaskID *uint64            `form:"subtask_id,omitempty" json:"subtask_id,omitempty" validate:"omitnil,min=0" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt time.Time          `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (ml *Vecstorelog) TableName() string {
	return "vecstorelogs"
}

// Valid is function to control input/output data
func (ml Vecstorelog) Valid() error {
	return validate.Struct(ml)
}

// Validate is function to use callback to control input/output data
func (ml Vecstorelog) Validate(db *gorm.DB) {
	if err := ml.Valid(); err != nil {
		db.AddError(err)
	}
}
