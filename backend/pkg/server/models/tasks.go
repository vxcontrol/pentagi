package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type TaskStatus string

const (
	TaskStatusCreated  TaskStatus = "created"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusWaiting  TaskStatus = "waiting"
	TaskStatusFinished TaskStatus = "finished"
	TaskStatusFailed   TaskStatus = "failed"
)

func (s TaskStatus) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s TaskStatus) Valid() error {
	switch s {
	case TaskStatusCreated,
		TaskStatusRunning,
		TaskStatusWaiting,
		TaskStatusFinished,
		TaskStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid TaskStatus: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s TaskStatus) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Task is model to contain task information
// nolint:lll
type Task struct {
	ID        uint64     `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Status    TaskStatus `form:"status" json:"status" validate:"valid,required" gorm:"type:TASK_STATUS;NOT NULL;default:'created'"`
	Title     string     `form:"title" json:"title" validate:"required" gorm:"type:TEXT;NOT NULL;default:'untitled'"`
	Input     string     `form:"input" json:"input" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Result    string     `form:"result" json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	FlowID    uint64     `form:"flow_id" json:"flow_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt time.Time  `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (t *Task) TableName() string {
	return "tasks"
}

// Valid is function to control input/output data
func (t Task) Valid() error {
	return validate.Struct(t)
}

// Validate is function to use callback to control input/output data
func (t Task) Validate(db *gorm.DB) {
	if err := t.Valid(); err != nil {
		db.AddError(err)
	}
}

// TaskSubtasks is model to contain task and linked subtasks information
// nolint:lll
type TaskSubtasks struct {
	Subtasks []Subtask `form:"subtasks" json:"subtasks" validate:"required" gorm:"foreignkey:TaskID;association_autoupdate:false;association_autocreate:false"`
	Task     `form:"" json:""`
}

// TableName returns the table name string to guaranty use correct table
func (ts *TaskSubtasks) TableName() string {
	return "tasks"
}

// Valid is function to control input/output data
func (ts TaskSubtasks) Valid() error {
	for i := range ts.Subtasks {
		if err := ts.Subtasks[i].Valid(); err != nil {
			return err
		}
	}
	return ts.Task.Valid()
}

// Validate is function to use callback to control input/output data
func (ts TaskSubtasks) Validate(db *gorm.DB) {
	if err := ts.Valid(); err != nil {
		db.AddError(err)
	}
}
