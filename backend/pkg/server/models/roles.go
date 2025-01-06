package models

import "github.com/jinzhu/gorm"

// Role is model to contain user role information
// nolint:lll
type Role struct {
	ID   uint64 `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Name string `form:"name" json:"name" validate:"max=50,required" gorm:"type:TEXT;NOT NULL;UNIQUE_INDEX"`
}

// TableName returns the table name string to guaranty use correct table
func (r *Role) TableName() string {
	return "roles"
}

// Valid is function to control input/output data
func (r Role) Valid() error {
	return validate.Struct(r)
}

// Validate is function to use callback to control input/output data
func (r Role) Validate(db *gorm.DB) {
	if err := r.Valid(); err != nil {
		db.AddError(err)
	}
}

// Privilege is model to contain user privileges
// nolint:lll
type Privilege struct {
	ID     uint64 `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	RoleID uint64 `form:"role_id" json:"role_id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL"`
	Name   string `form:"name" json:"name" validate:"max=70,required" gorm:"type:TEXT;NOT NULL"`
}

// TableName returns the table name string to guaranty use correct table
func (p *Privilege) TableName() string {
	return "privileges"
}

// Valid is function to control input/output data
func (p Privilege) Valid() error {
	return validate.Struct(p)
}

// RolePrivileges is model to contain user role privileges
// nolint:lll
type RolePrivileges struct {
	Privileges []Privilege `form:"privileges" json:"privileges" validate:"required" gorm:"foreignkey:RoleID;association_autoupdate:false;association_autocreate:false"`
	Role       `form:"" json:""`
}

// TableName returns the table name string to guaranty use correct table
func (rp *RolePrivileges) TableName() string {
	return "roles"
}

// Valid is function to control input/output data
func (rp RolePrivileges) Valid() error {
	for i := range rp.Privileges {
		if err := rp.Privileges[i].Valid(); err != nil {
			return err
		}
	}
	return rp.Role.Valid()
}

// Validate is function to use callback to control input/output data
func (rp RolePrivileges) Validate(db *gorm.DB) {
	if err := rp.Valid(); err != nil {
		db.AddError(err)
	}
}
