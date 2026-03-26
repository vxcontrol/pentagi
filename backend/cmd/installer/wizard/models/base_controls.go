package models

import (
	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/textinput"
)

func NewBooleanInput(styles styles.Styles, window window.Window, envVar loader.EnvVar) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.PlaceholderStyle = styles.FormPlaceholder
	input.ShowSuggestions = true
	input.SetSuggestions([]string{"true", "false"})

	if envVar.Default != "" {
		input.Placeholder = envVar.Default
	}

	if envVar.IsPresent() || envVar.IsChanged {
		input.SetValue(envVar.Value)
	}

	return input
}

func NewTextInput(styles styles.Styles, window window.Window, envVar loader.EnvVar) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.PlaceholderStyle = styles.FormPlaceholder

	if envVar.Default != "" {
		input.Placeholder = envVar.Default
	}

	if envVar.IsPresent() || envVar.IsChanged {
		input.SetValue(envVar.Value)
	}

	return input
}
