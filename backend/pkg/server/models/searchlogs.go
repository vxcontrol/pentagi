package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type SearchEngineType string

const (
	SearchEngineTypeGoogle     SearchEngineType = "google"
	SearchEngineTypeDuckduckgo SearchEngineType = "duckduckgo"
	SearchEngineTypeTavily     SearchEngineType = "tavily"
	SearchEngineTypeTraversaal SearchEngineType = "traversaal"
	SearchEngineTypePerplexity SearchEngineType = "perplexity"
	SearchEngineTypeBrowser    SearchEngineType = "browser"
	SearchEngineTypeSploitus   SearchEngineType = "sploitus"
)

func (s SearchEngineType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s SearchEngineType) Valid() error {
	switch s {
	case SearchEngineTypeGoogle,
		SearchEngineTypeDuckduckgo,
		SearchEngineTypeTavily,
		SearchEngineTypeTraversaal,
		SearchEngineTypePerplexity,
		SearchEngineTypeBrowser,
		SearchEngineTypeSploitus:
		return nil
	default:
		return fmt.Errorf("invalid SearchEngineType: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s SearchEngineType) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// Searchlog is model to contain search action information in the internet or local network
// nolint:lll
type Searchlog struct {
	ID        uint64           `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Initiator MsgchainType     `json:"initiator" validate:"valid,required" gorm:"type:MSGCHAIN_TYPE;NOT NULL"`
	Executor  MsgchainType     `json:"executor" validate:"valid,required" gorm:"type:MSGCHAIN_TYPE;NOT NULL"`
	Engine    SearchEngineType `json:"engine" validate:"valid,required" gorm:"type:SEARCHENGINE_TYPE;NOT NULL"`
	Query     string           `json:"query" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Result    string           `json:"result" validate:"omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	FlowID    uint64           `form:"flow_id" json:"flow_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL"`
	TaskID    *uint64          `form:"task_id,omitempty" json:"task_id,omitempty" validate:"omitnil,min=0" gorm:"type:BIGINT;NOT NULL"`
	SubtaskID *uint64          `form:"subtask_id,omitempty" json:"subtask_id,omitempty" validate:"omitnil,min=0" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt time.Time        `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (ml *Searchlog) TableName() string {
	return "searchlogs"
}

// Valid is function to control input/output data
func (ml Searchlog) Valid() error {
	return validate.Struct(ml)
}

// Validate is function to use callback to control input/output data
func (ml Searchlog) Validate(db *gorm.DB) {
	if err := ml.Valid(); err != nil {
		db.AddError(err)
	}
}
