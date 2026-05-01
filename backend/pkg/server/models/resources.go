package models

import "time"

// UserResource is the GORM model for a user-owned resource (file or directory)
// whose metadata lives in PostgreSQL and whose content (for files) is stored as
// a .blob file on disk keyed by MD5 hash.
//
// nolint:lll
type UserResource struct {
	ID        uint64    `form:"id"         json:"id"         gorm:"type:BIGINT;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	UserID    uint64    `form:"user_id"    json:"user_id"    gorm:"type:BIGINT;NOT NULL;index"`
	Hash      string    `form:"hash"       json:"hash"       gorm:"type:TEXT;NOT NULL;default:''"`
	Name      string    `form:"name"       json:"name"       gorm:"type:TEXT;NOT NULL"`
	Path      string    `form:"path"       json:"path"       gorm:"type:TEXT;NOT NULL"`
	Size      int64     `form:"size"       json:"size"       gorm:"type:BIGINT;NOT NULL;default:0"`
	IsDir     bool      `form:"is_dir"     json:"is_dir"     gorm:"type:BOOLEAN;NOT NULL;default:false"`
	CreatedAt time.Time `form:"created_at" json:"created_at"`
	UpdatedAt time.Time `form:"updated_at" json:"updated_at"`
}

// ResourceEntry is the REST response representation of a single resource.
type ResourceEntry struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	IsDir     bool      `json:"is_dir"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceList is the REST list response for resource operations.
type ResourceList struct {
	Items []ResourceEntry `json:"items"`
	Total uint64          `json:"total"`
}

// MkdirResourceRequest is the request body for creating a virtual directory.
type MkdirResourceRequest struct {
	Path string `json:"path" binding:"required"`
}

// MoveResourceRequest is the request body for moving or renaming a resource.
type MoveResourceRequest struct {
	Source      string `json:"source"      binding:"required"`
	Destination string `json:"destination" binding:"required"`
	Force       bool   `json:"force"`
}

// CopyResourceRequest is the request body for copying a resource.
type CopyResourceRequest struct {
	Source      string `json:"source"      binding:"required"`
	Destination string `json:"destination" binding:"required"`
	Force       bool   `json:"force"`
}
