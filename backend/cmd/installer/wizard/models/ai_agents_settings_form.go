package models

import (
	"fmt"
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// AIAgentsSettingsFormModel represents the AI agents settings form
type AIAgentsSettingsFormModel struct {
	*BaseScreen
}

// NewAIAgentsSettingsFormModel creates a new AI agents settings form model
func NewAIAgentsSettingsFormModel(c controller.Controller, s styles.Styles, w window.Window) *AIAgentsSettingsFormModel {
	m := &AIAgentsSettingsFormModel{}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BaseScreenHandler interface implementation

func (m *AIAgentsSettingsFormModel) BuildForm() tea.Cmd {
	cfg := m.GetController().GetAIAgentsConfig()

	askUserInput := NewBooleanInput(m.GetStyles(), m.GetWindow(), cfg.HumanInTheLoop)
	useAgentsInput := NewBooleanInput(m.GetStyles(), m.GetWindow(), cfg.AssistantUseAgents)

	fields := []FormField{
		{
			Key:         "ask_user",
			Title:       locale.ToolsAIAgentsSettingHumanInTheLoop,
			Description: locale.ToolsAIAgentsSettingHumanInTheLoopDesc,
			Required:    false,
			Masked:      false,
			Input:       askUserInput,
			Value:       askUserInput.Value(),
			Suggestions: askUserInput.AvailableSuggestions(),
		},
		{
			Key:         "assistant_use_agents",
			Title:       locale.ToolsAIAgentsSettingUseAgents,
			Description: locale.ToolsAIAgentsSettingUseAgentsDesc,
			Required:    false,
			Masked:      false,
			Input:       useAgentsInput,
			Value:       useAgentsInput.Value(),
			Suggestions: useAgentsInput.AvailableSuggestions(),
		},
	}

	m.SetFormFields(fields)
	return askUserInput.Focus()
}

func (m *AIAgentsSettingsFormModel) GetFormTitle() string {
	return locale.ToolsAIAgentsSettingsFormTitle
}
func (m *AIAgentsSettingsFormModel) GetFormDescription() string {
	return locale.ToolsAIAgentsSettingsFormDescription
}
func (m *AIAgentsSettingsFormModel) GetFormName() string    { return locale.ToolsAIAgentsSettingsFormName }
func (m *AIAgentsSettingsFormModel) GetFormSummary() string { return "" }

func (m *AIAgentsSettingsFormModel) GetFormOverview() string {
	var sections []string
	sections = append(sections, m.styles.Subtitle.Render(locale.ToolsAIAgentsSettingsFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.styles.Paragraph.Bold(true).Render(locale.ToolsAIAgentsSettingsFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.styles.Paragraph.Render(locale.ToolsAIAgentsSettingsFormOverview))
	return strings.Join(sections, "\n")
}

func (m *AIAgentsSettingsFormModel) GetCurrentConfiguration() string {
	sections := []string{m.GetStyles().Subtitle.Render(m.GetFormName())}
	cfg := m.GetController().GetAIAgentsConfig()

	// ask_user
	if val := cfg.HumanInTheLoop.Value; val == "true" || (val == "" && cfg.HumanInTheLoop.Default == "true") {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ToolsAIAgentsSettingHumanInTheLoop, m.styles.Success.Render(locale.StatusEnabled)))
	} else {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ToolsAIAgentsSettingHumanInTheLoop, m.styles.Warning.Render(locale.StatusDisabled)))
	}

	// assistant_use_agents
	if val := cfg.AssistantUseAgents.Value; val == "true" || (val == "" && cfg.AssistantUseAgents.Default == "true") {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ToolsAIAgentsSettingUseAgents, m.styles.Success.Render(locale.StatusEnabled)))
	} else {
		sections = append(sections, fmt.Sprintf("• %s: %s", locale.ToolsAIAgentsSettingUseAgents, m.styles.Warning.Render(locale.StatusDisabled)))
	}

	return strings.Join(sections, "\n")
}

func (m *AIAgentsSettingsFormModel) IsConfigured() bool {
	cfg := m.GetController().GetAIAgentsConfig()
	return cfg.HumanInTheLoop.IsPresent() || cfg.HumanInTheLoop.IsChanged || cfg.AssistantUseAgents.IsPresent() || cfg.AssistantUseAgents.IsChanged
}

func (m *AIAgentsSettingsFormModel) GetHelpContent() string {
	var sections []string
	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ToolsAIAgentsSettingsFormTitle))
	sections = append(sections, "")
	sections = append(sections, locale.ToolsAIAgentsSettingsHelp)
	return strings.Join(sections, "\n")
}

func (m *AIAgentsSettingsFormModel) HandleSave() error {
	fields := m.GetFormFields()
	if len(fields) != 2 {
		return fmt.Errorf("unexpected number of fields: %d", len(fields))
	}

	// normalize and validate booleans
	normalize := func(v string) (string, error) {
		vv := strings.ToLower(strings.TrimSpace(v))
		if vv != "" && vv != "true" && vv != "false" {
			return "", fmt.Errorf("invalid boolean value: %s", v)
		}
		return vv, nil
	}

	askUserVal, err := normalize(fields[0].Input.Value())
	if err != nil {
		return err
	}
	useAgentsVal, err := normalize(fields[1].Input.Value())
	if err != nil {
		return err
	}

	cur := m.GetController().GetAIAgentsConfig()
	newCfg := &controller.AIAgentsConfig{
		HumanInTheLoop:     cur.HumanInTheLoop,
		AssistantUseAgents: cur.AssistantUseAgents,
	}
	newCfg.HumanInTheLoop.Value = askUserVal
	newCfg.AssistantUseAgents.Value = useAgentsVal

	if err := m.GetController().UpdateAIAgentsConfig(newCfg); err != nil {
		return fmt.Errorf("error setting config: %v", err)
	}

	logger.Log("[AIAgentsSettingsFormModel] SAVE: success")
	return nil
}

func (m *AIAgentsSettingsFormModel) HandleReset() {
	cfg := m.GetController().ResetAIAgentsConfig()
	fields := m.GetFormFields()
	if len(fields) >= 1 {
		fields[0].Input.SetValue(cfg.HumanInTheLoop.Value)
		fields[0].Value = fields[0].Input.Value()
	}
	if len(fields) >= 2 {
		fields[1].Input.SetValue(cfg.AssistantUseAgents.Value)
		fields[1].Value = fields[1].Input.Value()
	}
	m.SetFormFields(fields)
}

func (m *AIAgentsSettingsFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {}
func (m *AIAgentsSettingsFormModel) GetFormFields() []FormField                               { return m.fields }
func (m *AIAgentsSettingsFormModel) SetFormFields(fields []FormField)                         { m.fields = fields }

// Update method - handle screen-specific input
func (m *AIAgentsSettingsFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}
	return m, m.BaseScreen.Update(msg)
}

// Compile-time interface validation
var _ BaseScreenModel = (*AIAgentsSettingsFormModel)(nil)
var _ BaseScreenHandler = (*AIAgentsSettingsFormModel)(nil)
