package models

// Settings is model to contain application settings information
// nolint:lll
type Settings struct {
	Debug              bool   `json:"debug" example:"false"`
	AskUser            bool   `json:"ask_user" example:"false"`
	Version            string `json:"version" example:"v1.0.0"`
	DockerInside       bool   `json:"docker_inside" example:"false"`
	IsDevelopMode      bool   `json:"is_develop_mode" example:"false"`
	AssistantUseAgents bool   `json:"assistant_use_agents" example:"false"`
}

// Valid is function to control input/output data
func (s Settings) Valid() error {
	return validate.Struct(s)
}

// DeafGuardConfig is the runtime command-classifier configuration.
type DeafGuardConfig struct {
	Enabled bool   `json:"enabled" example:"true"`
	Mode    string `json:"mode" example:"log"`
}

// DeafGuardConfigUpdate is a partial update for Deaf Guard configuration.
type DeafGuardConfigUpdate struct {
	Enabled *bool   `json:"enabled"`
	Mode    *string `json:"mode"`
}
