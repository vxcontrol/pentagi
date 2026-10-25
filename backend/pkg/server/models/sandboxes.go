package models

// CreateSandbox is a request to provision a headless Docker sandbox: a
// container with no agent loop and no LLM provider behind it, driven entirely
// by whoever holds the API token.
type CreateSandbox struct {
	Title string `form:"title,omitempty" json:"title,omitempty" validate:"omitempty,max=70" example:"ad-hoc recon"`
	Image string `form:"image,omitempty" json:"image,omitempty" validate:"omitempty,max=255" example:"vxcontrol/kali-linux"`
}

// Valid is function to control input/output data
func (cs CreateSandbox) Valid() error {
	return validate.Struct(cs)
}

// SandboxExec is one command run inside a sandbox's container.
type SandboxExec struct {
	Input   string `json:"input" validate:"required" example:"nmap -sV 10.0.0.5"`
	Cwd     string `json:"cwd,omitempty" validate:"omitempty"`
	Detach  bool   `json:"detach,omitempty" example:"false"`
	Timeout int64  `json:"timeout,omitempty" validate:"omitempty,min=0,max=10800" example:"60"`
}

// Valid is function to control input/output data
func (se SandboxExec) Valid() error {
	return validate.Struct(se)
}

// SandboxFileWrite overwrites a file inside a sandbox's container.
type SandboxFileWrite struct {
	Path    string `json:"path" validate:"required" example:"/work/notes.txt"`
	Content string `json:"content"`
}

// Valid is function to control input/output data
func (sfw SandboxFileWrite) Valid() error {
	return validate.Struct(sfw)
}
