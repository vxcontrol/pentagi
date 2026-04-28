package models

import "time"

// FlowFile represents a single entry in the flow's local file cache.
type FlowFile struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"` // relative: "uploads/<x>" or resources/<x> or container/<x>"
	Size       int64     `json:"size"`
	IsDir      bool      `json:"isDir"`
	ModifiedAt time.Time `json:"modifiedAt"`
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
	IsDir      bool      `json:"isDir"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// ContainerFiles is the list response for container directory listing.
type ContainerFiles struct {
	Path  string          `json:"path"`
	Files []ContainerFile `json:"files"`
	Total uint64          `json:"total"`
}

// PullFlowFilesRequest is the request body for pulling files from a container.
type PullFlowFilesRequest struct {
	// Path is an arbitrary path inside the container, e.g. "/etc/nginx/conf" or "/work/uploads/report.txt".
	Path string `json:"path"`
	// Force overwrites the local cache entry if it already exists.
	Force bool `json:"force"`
}

// AddResourcesRequest is the request body for copying user resources into a flow.
type AddResourcesRequest struct {
	// IDs is the list of user resource IDs to copy into the flow resources directory.
	IDs []string `json:"ids" binding:"required,min=1"`
	// Force overwrites files that already exist in the flow resources directory.
	Force bool `json:"force"`
}

// AddResourceFromFlowRequest is the request body for promoting a flow file to user resources.
type AddResourceFromFlowRequest struct {
	// SourcePath is a relative path within the flow cache, e.g. "container/work/result.md" or "uploads/task.md".
	SourcePath string `json:"sourcePath" binding:"required"`
	// Destination is the virtual path the resource will have in the user's resource tree.
	Destination string `json:"destination" binding:"required"`
	// Force overwrites an existing resource at Destination if one already exists.
	Force bool `json:"force"`
}
