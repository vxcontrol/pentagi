package models

import "time"

// FlowFile represents a single entry in the flow's local file cache.
type FlowFile struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"` // relative: "uploads/<x>" or resources/<x> or container/<x>"
	Size       int64     `json:"size"`
	IsDir      bool      `json:"is_dir"`
	ModifiedAt time.Time `json:"modified_at"`
}

// FlowFiles is the list response for flow file operations.
type FlowFiles struct {
	Files []FlowFile `json:"files"`
	Total uint64     `json:"total"`
}

// ContainerFile represents a single entry in the running container's filesystem.
type ContainerFile struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	IsDir      bool      `json:"is_dir"`
	ModifiedAt time.Time `json:"modified_at"`
}

// ContainerFiles is the list response for container directory listing.
type ContainerFiles struct {
	Path  string          `json:"path"`
	Files []ContainerFile `json:"files"`
	Total uint64          `json:"total"`
}

// PullFlowFilesRequest is the request body for pulling files from a container.
type PullFlowFilesRequest struct {
	// Path is a single container path (maintained for backward compatibility).
	// Combined with Paths if both are provided.
	Path string `json:"path"`
	// Paths is a list of container paths to pull. Combined with Path if provided.
	Paths []string `json:"paths"`
	// Force overwrites local cache entries that already exist.
	Force bool `json:"force"`
}

// AddResourcesRequest is the request body for copying user resources into a flow.
type AddResourcesRequest struct {
	// IDs is the list of user resource IDs to copy into the flow resources directory.
	IDs []uint64 `json:"ids" binding:"required,min=1"`
	// Force overwrites files that already exist in the flow resources directory.
	Force bool `json:"force"`
}

// AddResourceFromFlowRequest is the request body for promoting one or more flow
// files / directories into the user's global resource store.
//
// At least one of Source or Sources must be non-empty after deduplication.
//
//   - Single source (Source XOR one entry in Sources):
//     Destination is the exact target path (file) or root directory (dir tree).
//   - Multiple sources:
//     Destination is treated as a base directory; each source's base name is
//     appended automatically, e.g. sources ["uploads/a.txt", "container/b.txt"]
//     with destination "results" → "results/a.txt" and "results/b.txt".
type AddResourceFromFlowRequest struct {
	// Source is a single relative path within the flow cache (kept for backward
	// compatibility). Combined with Sources when both are provided.
	Source string `json:"source"`
	// Sources is a list of relative paths within the flow cache. Combined with
	// Source when both are provided; duplicate entries are silently removed.
	Sources []string `json:"sources"`
	// Destination is the virtual path prefix in the user's resource tree.
	Destination string `json:"destination" binding:"required"`
	// Force overwrites existing resources at the target paths when true.
	Force bool `json:"force"`
}
