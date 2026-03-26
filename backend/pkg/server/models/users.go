package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

const RoleUser = 2

type UserStatus string

const (
	UserStatusCreated UserStatus = "created"
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
)

func (s UserStatus) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s UserStatus) Valid() error {
	switch s {
	case UserStatusCreated, UserStatusActive, UserStatusBlocked:
		return nil
	default:
		return fmt.Errorf("invalid UserStatus: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s UserStatus) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

type UserType string

const (
	UserTypeLocal UserType = "local"
	UserTypeOAuth UserType = "oauth"
	UserTypeAPI   UserType = "api"
)

func (s UserType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (s UserType) Valid() error {
	switch s {
	case UserTypeLocal, UserTypeOAuth, UserTypeAPI:
		return nil
	default:
		return fmt.Errorf("invalid UserType: %s", s)
	}
}

// Validate is function to use callback to control input/output data
func (s UserType) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// User is model to contain user information
// nolint:lll
type User struct {
	ID                     uint64     `form:"id" json:"id" validate:"min=0,numeric" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Hash                   string     `form:"hash" json:"hash" validate:"len=32,hexadecimal,lowercase,omitempty" gorm:"type:TEXT;NOT NULL;UNIQUE_INDEX;default:MD5(RANDOM()::text)"`
	Type                   UserType   `form:"type" json:"type" validate:"valid,required" gorm:"type:USER_TYPE;NOT NULL;default:'local'"`
	Mail                   string     `form:"mail" json:"mail" validate:"max=50,vmail,required" gorm:"type:TEXT;NOT NULL;UNIQUE_INDEX"`
	Name                   string     `form:"name" json:"name" validate:"max=70,omitempty" gorm:"type:TEXT;NOT NULL;default:''"`
	Status                 UserStatus `form:"status" json:"status" validate:"valid,required" gorm:"type:USER_STATUS;NOT NULL;default:'created'"`
	RoleID                 uint64     `form:"role_id" json:"role_id" validate:"min=0,numeric,required" gorm:"type:BIGINT;NOT NULL;default:2"`
	PasswordChangeRequired bool       `form:"password_change_required" json:"password_change_required" gorm:"type:BOOL;NOT NULL;default:false"`
	Provider               *string    `form:"provider,omitempty" json:"provider,omitempty" validate:"omitempty" gorm:"type:TEXT"`
	CreatedAt              time.Time  `form:"created_at" json:"created_at" validate:"omitempty" gorm:"type:TIMESTAMPTZ;NOT NULL;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (u *User) TableName() string {
	return "users"
}

// Valid is function to control input/output data
func (u User) Valid() error {
	return validate.Struct(u)
}

// Validate is function to use callback to control input/output data
func (u User) Validate(db *gorm.DB) {
	if err := u.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserPassword is model to contain user information
type UserPassword struct {
	Password string `form:"password" json:"password" validate:"max=100,required" gorm:"column:password;type:TEXT"`
	User     `form:"" json:""`
}

// TableName returns the table name string to guaranty use correct table
func (up *UserPassword) TableName() string {
	return "users"
}

// Valid is function to control input/output data
func (up UserPassword) Valid() error {
	if err := up.User.Valid(); err != nil {
		return err
	}
	return validate.Struct(up)
}

// Validate is function to use callback to control input/output data
func (up UserPassword) Validate(db *gorm.DB) {
	if err := up.Valid(); err != nil {
		db.AddError(err)
	}
}

// Login is model to contain user information on Login procedure
// nolint:lll
type Login struct {
	Mail     string `form:"mail" json:"mail" validate:"max=50,required" gorm:"type:TEXT;NOT NULL;UNIQUE_INDEX"`
	Password string `form:"password" json:"password" validate:"min=4,max=100,required" gorm:"type:TEXT"`
}

// TableName returns the table name string to guaranty use correct table
func (sin *Login) TableName() string {
	return "users"
}

// Valid is function to control input/output data
func (sin Login) Valid() error {
	return validate.Struct(sin)
}

// Validate is function to use callback to control input/output data
func (sin Login) Validate(db *gorm.DB) {
	if err := sin.Valid(); err != nil {
		db.AddError(err)
	}
}

// AuthCallback is model to contain auth data information from external OAuth application
type AuthCallback struct {
	Code    string `form:"code" json:"code" validate:"required"`
	IdToken string `form:"id_token" json:"id_token" validate:"required,jwt"`
	Scope   string `form:"scope" json:"scope" validate:"required,oauth_min_scope"`
	State   string `form:"state" json:"state" validate:"required"`
}

// Valid is function to control input/output data
func (au AuthCallback) Valid() error {
	return validate.Struct(au)
}

// Password is model to contain user password to change it
// nolint:lll
type Password struct {
	CurrentPassword string `form:"current_password" json:"current_password" validate:"nefield=Password,min=5,max=100,required" gorm:"-"`
	Password        string `form:"password" json:"password" validate:"stpass,max=100,required" gorm:"type:TEXT"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password" validate:"eqfield=Password" gorm:"-"`
}

// TableName returns the table name string to guaranty use correct table
func (p *Password) TableName() string {
	return "users"
}

// Valid is function to control input/output data
func (p Password) Valid() error {
	return validate.Struct(p)
}

// Validate is function to use callback to control input/output data
func (p Password) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserRole is model to contain user information linked with user role
// nolint:lll
type UserRole struct {
	Role Role `form:"role,omitempty" json:"role,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	User `form:"" json:""`
}

// Valid is function to control input/output data
func (ur UserRole) Valid() error {
	if err := ur.Role.Valid(); err != nil {
		return err
	}
	return ur.User.Valid()
}

// Validate is function to use callback to control input/output data
func (ur UserRole) Validate(db *gorm.DB) {
	if err := ur.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserRole is model to contain user information linked with user role
// nolint:lll
type UserRolePrivileges struct {
	Role RolePrivileges `form:"role,omitempty" json:"role,omitempty" gorm:"association_autoupdate:false;association_autocreate:false"`
	User `form:"" json:""`
}

// Valid is function to control input/output data
func (urp UserRolePrivileges) Valid() error {
	if err := urp.Role.Valid(); err != nil {
		return err
	}
	return urp.User.Valid()
}

// Validate is function to use callback to control input/output data
func (urp UserRolePrivileges) Validate(db *gorm.DB) {
	if err := urp.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserPreferencesOptions is model to contain user preferences as JSON
type UserPreferencesOptions struct {
	FavoriteFlows []int64 `json:"favoriteFlows"`
}

// Value implements driver.Valuer interface for database write
func (upo UserPreferencesOptions) Value() (driver.Value, error) {
	return json.Marshal(upo)
}

// Scan implements sql.Scanner interface for database read
func (upo *UserPreferencesOptions) Scan(value any) error {
	if value == nil {
		*upo = UserPreferencesOptions{FavoriteFlows: []int64{}}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan UserPreferencesOptions: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, upo)
}

// UserPreferences is model to contain user preferences information
type UserPreferences struct {
	ID          uint64                 `json:"id" gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	UserID      uint64                 `json:"user_id" gorm:"type:BIGINT;NOT NULL;UNIQUE_INDEX"`
	Preferences UserPreferencesOptions `json:"preferences" gorm:"type:JSONB;NOT NULL"`
	CreatedAt   time.Time              `json:"created_at" gorm:"type:TIMESTAMPTZ;NOT NULL;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"type:TIMESTAMPTZ;NOT NULL;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (up *UserPreferences) TableName() string {
	return "user_preferences"
}

// Valid is function to control input/output data
func (up UserPreferences) Valid() error {
	if up.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

// Validate is function to use callback to control input/output data
func (up UserPreferences) Validate(db *gorm.DB) {
	if err := up.Valid(); err != nil {
		db.AddError(err)
	}
}

// NewUserPreferences creates a new UserPreferences with default values
func NewUserPreferences(userID uint64) *UserPreferences {
	return &UserPreferences{
		UserID: userID,
		Preferences: UserPreferencesOptions{
			FavoriteFlows: []int64{},
		},
	}
}

// UserWithPreferences is model to combine User and UserPreferences for transactional creation
type UserWithPreferences struct {
	User        User
	Preferences UserPreferences
}
