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
//
// At least one of Source or Sources must be non-empty after deduplication.
//
//   - Single source: Destination is the exact target path (file) or root
//     directory (dir tree), identical to the original single-source behaviour.
//   - Multiple sources: Destination is treated as a base directory; each
//     source's base name is appended automatically, e.g. moving "a.txt" and
//     "b.txt" to "archive" produces "archive/a.txt" and "archive/b.txt".
//     All moves happen atomically inside a single DB transaction.
type MoveResourceRequest struct {
	// Source is kept for backward compatibility (single source).
	// Combined with Sources when both are provided; duplicates are removed.
	Source string `json:"source"`
	// Sources is a list of virtual resource paths to move.
	// Combined with Source when both are provided; duplicates are removed.
	Sources []string `json:"sources"`
	// Destination is the exact target path (single source) or base directory
	// (multiple sources). Required.
	Destination string `json:"destination" binding:"required"`
	// Force overwrites existing resources at the target paths when true.
	Force bool `json:"force"`
}

// CopyResourceRequest is the request body for copying a resource.
//
// At least one of Source or Sources must be non-empty after deduplication.
//
//   - Single source: Destination is the exact target path (file) or root
//     directory (dir tree), identical to the original single-source behaviour.
//   - Multiple sources: Destination is treated as a base directory; each
//     source's base name is appended automatically, e.g. copying "a.txt" and
//     "b.txt" to "backup" produces "backup/a.txt" and "backup/b.txt".
//     All copies happen atomically inside a single DB transaction.
type CopyResourceRequest struct {
	// Source is kept for backward compatibility (single source).
	// Combined with Sources when both are provided; duplicates are removed.
	Source string `json:"source"`
	// Sources is a list of virtual resource paths to copy.
	// Combined with Source when both are provided; duplicates are removed.
	Sources []string `json:"sources"`
	// Destination is the exact target path (single source) or base directory
	// (multiple sources). Required.
	Destination string `json:"destination" binding:"required"`
	// Force overwrites existing resources at the target paths when true.
	Force bool `json:"force"`
}
