package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type TermlogType string

const (
	TermlogTypeStdin  TermlogType = "stdin"
	TermlogTypeStdout TermlogType = "stdout"
	TermlogTypeStderr TermlogType = "stderr"
)

func (s TermlogType) String() string {
	return string(s)
}

// Termlog is model to contain termlog information
// nolint:lll
type Termlog struct {
	ID          uint64      `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type        TermlogType `form:"type" json:"type" validate:"oneof=stdin stdout stderr,required" gorm:"type:TERMLOG_TYPE;NOT NULL"`
	Text        string      `form:"text" json:"text" validate:"required" gorm:"type:TEXT;NOT NULL"`
	ContainerID uint64      `form:"container_id" json:"container_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt   time.Time   `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (tl *Termlog) TableName() string {
	return "termlogs"
}

// Valid is function to control input/output data
func (tl Termlog) Valid() error {
	return validate.Struct(tl)
}

// Validate is function to use callback to control input/output data
func (tl Termlog) Validate(db *gorm.DB) {
	if err := tl.Valid(); err != nil {
		db.AddError(err)
	}
}
