package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type ContainerStatus string

const (
	ContainerStatusStarting ContainerStatus = "starting"
	ContainerStatusRunning  ContainerStatus = "running"
	ContainerStatusStopped  ContainerStatus = "stopped"
	ContainerStatusDeleted  ContainerStatus = "deleted"
	ContainerStatusFailed   ContainerStatus = "failed"
)

func (s ContainerStatus) String() string {
	return string(s)
}

type ContainerType string

const (
	ContainerTypePrimary   ContainerType = "primary"
	ContainerTypeSecondary ContainerType = "secondary"
)

func (t ContainerType) String() string {
	return string(t)
}

// Container is model to contain container information
// nolint:lll
type Container struct {
	ID        uint64          `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Type      ContainerType   `form:"type" json:"type" validate:"oneof=primary secondary,required" gorm:"type:CONTAINER_TYPE;NOT NULL;default:'primary'"`
	Name      string          `form:"name" json:"name" validate:"required" gorm:"type:TEXT;NOT NULL;default:MD5(RANDOM()::text)"`
	Image     string          `form:"image" json:"image" validate:"required" gorm:"type:TEXT;NOT NULL"`
	Status    ContainerStatus `form:"status" json:"status" validate:"oneof=starting running stopped deleted failed,required" gorm:"type:CONTAINER_STATUS;NOT NULL;default:'starting'"`
	LocalID   string          `form:"local_id" json:"local_id" validate:"required" gorm:"type:TEXT;NOT NULL"`
	LocalDir  string          `form:"local_dir" json:"local_dir" validate:"required" gorm:"type:TEXT;NOT NULL"`
	FlowID    uint64          `form:"flow_id" json:"flow_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	CreatedAt time.Time       `form:"created_at,omitempty" json:"created_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time       `form:"updated_at,omitempty" json:"updated_at,omitempty" validate:"omitempty" gorm:"type:TIMESTAMPTZ;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (c *Container) TableName() string {
	return "containers"
}

// Valid is function to control input/output data
func (c Container) Valid() error {
	return validate.Struct(c)
}

// Validate is function to use callback to control input/output data
func (c Container) Validate(db *gorm.DB) {
	if err := c.Valid(); err != nil {
		db.AddError(err)
	}
}
