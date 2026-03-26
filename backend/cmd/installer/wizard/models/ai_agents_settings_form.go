package models

import (
	"fmt"
	"strconv"
	"strings"

	"pentagi/cmd/installer/loader"
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

	fields := []FormField{
		m.createBooleanField(
			"ask_user",
			locale.ToolsAIAgentsSettingHumanInTheLoop,
			locale.ToolsAIAgentsSettingHumanInTheLoopDesc,
			cfg.HumanInTheLoop,
		),
		m.createBooleanField(
			"assistant_use_agents",
			locale.ToolsAIAgentsSettingUseAgents,
			locale.ToolsAIAgentsSettingUseAgentsDesc,
			cfg.AssistantUseAgents,
		),
		m.createBooleanField(
			"execution_monitor_enabled",
			locale.ToolsAIAgentsSettingExecutionMonitor,
			locale.ToolsAIAgentsSettingExecutionMonitorDesc,
			cfg.ExecutionMonitorEnabled,
		),
		m.createIntegerField(
			"execution_monitor_same_tool_limit",
			locale.ToolsAIAgentsSettingSameToolLimit,
			locale.ToolsAIAgentsSettingSameToolLimitDesc,
			cfg.ExecutionMonitorSameToolLimit,
			1,
			50,
		),
		m.createIntegerField(
			"execution_monitor_total_tool_limit",
			locale.ToolsAIAgentsSettingTotalToolLimit,
			locale.ToolsAIAgentsSettingTotalToolLimitDesc,
			cfg.ExecutionMonitorTotalToolLimit,
			1,
			100,
		),
		m.createIntegerField(
			"max_general_agent_tool_calls",
			locale.ToolsAIAgentsSettingMaxGeneralToolCalls,
			locale.ToolsAIAgentsSettingMaxGeneralToolCallsDesc,
			cfg.MaxGeneralAgentToolCalls,
			10,
			500,
		),
		m.createIntegerField(
			"max_limited_agent_tool_calls",
			locale.ToolsAIAgentsSettingMaxLimitedToolCalls,
			locale.ToolsAIAgentsSettingMaxLimitedToolCallsDesc,
			cfg.MaxLimitedAgentToolCalls,
			5,
			200,
		),
		m.createBooleanField(
			"agent_planning_step_enabled",
			locale.ToolsAIAgentsSettingTaskPlanning,
			locale.ToolsAIAgentsSettingTaskPlanningDesc,
			cfg.AgentPlanningStepEnabled,
		),
	}

	m.SetFormFields(fields)
	return fields[0].Input.Focus()
}

func (m *AIAgentsSettingsFormModel) createBooleanField(key, title, description string, envVar loader.EnvVar) FormField {
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), envVar)

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
		Suggestions: input.AvailableSuggestions(),
	}
}

func (m *AIAgentsSettingsFormModel) createIntegerField(key, title, description string, envVar loader.EnvVar, min, max int) FormField {
	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)

	// set placeholder with range info
	if envVar.Default != "" {
		input.Placeholder = fmt.Sprintf("%s (%d-%s)", envVar.Default, min, m.formatNumber(max))
	} else {
		input.Placeholder = fmt.Sprintf("(%d-%s)", min, m.formatNumber(max))
	}

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      false,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *AIAgentsSettingsFormModel) validateBooleanField(value, fieldName string) error {
	if value != "" && value != "true" && value != "false" {
		return fmt.Errorf("invalid boolean value for %s: %s (must be 'true' or 'false')", fieldName, value)
	}
	return nil
}

func (m *AIAgentsSettingsFormModel) validateIntegerField(value, fieldName string, min, max int) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("%s cannot be empty", fieldName)
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %s", fieldName, value)
	}

	if intVal < min || intVal > max {
		return 0, fmt.Errorf("%s must be between %d and %s", fieldName, min, m.formatNumber(max))
	}

	return intVal, nil
}

func (m *AIAgentsSettingsFormModel) formatNumber(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
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

	// helper function for boolean fields
	displayBoolean := func(envVar loader.EnvVar, label string) {
		val := envVar.Value
		if val == "" {
			val = envVar.Default
		}
		if val == "true" {
			sections = append(sections, fmt.Sprintf("• %s: %s", label, m.styles.Success.Render(locale.StatusEnabled)))
		} else {
			sections = append(sections, fmt.Sprintf("• %s: %s", label, m.styles.Warning.Render(locale.StatusDisabled)))
		}
	}

	// helper function for integer fields
	displayInteger := func(envVar loader.EnvVar, label string) {
		val := envVar.Value
		if val == "" {
			val = envVar.Default
		}
		if val != "" {
			sections = append(sections, fmt.Sprintf("• %s: %s", label, m.styles.Info.Render(val)))
		} else {
			sections = append(sections, fmt.Sprintf("• %s: %s", label, m.styles.Warning.Render("not set")))
		}
	}

	// basic settings
	displayBoolean(cfg.HumanInTheLoop, locale.ToolsAIAgentsSettingHumanInTheLoop)
	displayBoolean(cfg.AssistantUseAgents, locale.ToolsAIAgentsSettingUseAgents)

	// execution monitoring
	displayBoolean(cfg.ExecutionMonitorEnabled, locale.ToolsAIAgentsSettingExecutionMonitor)
	displayInteger(cfg.ExecutionMonitorSameToolLimit, locale.ToolsAIAgentsSettingSameToolLimit)
	displayInteger(cfg.ExecutionMonitorTotalToolLimit, locale.ToolsAIAgentsSettingTotalToolLimit)

	// tool call limits
	displayInteger(cfg.MaxGeneralAgentToolCalls, locale.ToolsAIAgentsSettingMaxGeneralToolCalls)
	displayInteger(cfg.MaxLimitedAgentToolCalls, locale.ToolsAIAgentsSettingMaxLimitedToolCalls)

	// task planning
	displayBoolean(cfg.AgentPlanningStepEnabled, locale.ToolsAIAgentsSettingTaskPlanning)

	return strings.Join(sections, "\n")
}

func (m *AIAgentsSettingsFormModel) IsConfigured() bool {
	cfg := m.GetController().GetAIAgentsConfig()
	return cfg.HumanInTheLoop.IsPresent() || cfg.HumanInTheLoop.IsChanged ||
		cfg.AssistantUseAgents.IsPresent() || cfg.AssistantUseAgents.IsChanged ||
		cfg.ExecutionMonitorEnabled.IsPresent() || cfg.ExecutionMonitorEnabled.IsChanged ||
		cfg.ExecutionMonitorSameToolLimit.IsPresent() || cfg.ExecutionMonitorSameToolLimit.IsChanged ||
		cfg.ExecutionMonitorTotalToolLimit.IsPresent() || cfg.ExecutionMonitorTotalToolLimit.IsChanged ||
		cfg.MaxGeneralAgentToolCalls.IsPresent() || cfg.MaxGeneralAgentToolCalls.IsChanged ||
		cfg.MaxLimitedAgentToolCalls.IsPresent() || cfg.MaxLimitedAgentToolCalls.IsChanged ||
		cfg.AgentPlanningStepEnabled.IsPresent() || cfg.AgentPlanningStepEnabled.IsChanged
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
	if len(fields) != 8 {
		return fmt.Errorf("unexpected number of fields: %d", len(fields))
	}

	cur := m.GetController().GetAIAgentsConfig()
	newCfg := &controller.AIAgentsConfig{
		HumanInTheLoop:                 cur.HumanInTheLoop,
		AssistantUseAgents:             cur.AssistantUseAgents,
		ExecutionMonitorEnabled:        cur.ExecutionMonitorEnabled,
		ExecutionMonitorSameToolLimit:  cur.ExecutionMonitorSameToolLimit,
		ExecutionMonitorTotalToolLimit: cur.ExecutionMonitorTotalToolLimit,
		MaxGeneralAgentToolCalls:       cur.MaxGeneralAgentToolCalls,
		MaxLimitedAgentToolCalls:       cur.MaxLimitedAgentToolCalls,
		AgentPlanningStepEnabled:       cur.AgentPlanningStepEnabled,
	}

	// validate and set each field
	for i, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "ask_user":
			if err := m.validateBooleanField(value, locale.ToolsAIAgentsSettingHumanInTheLoop); err != nil {
				return err
			}
			newCfg.HumanInTheLoop.Value = value

		case "assistant_use_agents":
			if err := m.validateBooleanField(value, locale.ToolsAIAgentsSettingUseAgents); err != nil {
				return err
			}
			newCfg.AssistantUseAgents.Value = value

		case "execution_monitor_enabled":
			if err := m.validateBooleanField(value, locale.ToolsAIAgentsSettingExecutionMonitor); err != nil {
				return err
			}
			newCfg.ExecutionMonitorEnabled.Value = value

		case "execution_monitor_same_tool_limit":
			if val, err := m.validateIntegerField(value, locale.ToolsAIAgentsSettingSameToolLimit, 1, 50); err != nil {
				return err
			} else {
				newCfg.ExecutionMonitorSameToolLimit.Value = strconv.Itoa(val)
			}

		case "execution_monitor_total_tool_limit":
			if val, err := m.validateIntegerField(value, locale.ToolsAIAgentsSettingTotalToolLimit, 1, 100); err != nil {
				return err
			} else {
				newCfg.ExecutionMonitorTotalToolLimit.Value = strconv.Itoa(val)
			}

		case "max_general_agent_tool_calls":
			if val, err := m.validateIntegerField(value, locale.ToolsAIAgentsSettingMaxGeneralToolCalls, 10, 500); err != nil {
				return err
			} else {
				newCfg.MaxGeneralAgentToolCalls.Value = strconv.Itoa(val)
			}

		case "max_limited_agent_tool_calls":
			if val, err := m.validateIntegerField(value, locale.ToolsAIAgentsSettingMaxLimitedToolCalls, 5, 200); err != nil {
				return err
			} else {
				newCfg.MaxLimitedAgentToolCalls.Value = strconv.Itoa(val)
			}

		case "agent_planning_step_enabled":
			if err := m.validateBooleanField(value, locale.ToolsAIAgentsSettingTaskPlanning); err != nil {
				return err
			}
			newCfg.AgentPlanningStepEnabled.Value = value

		default:
			return fmt.Errorf("unknown field key at index %d: %s", i, field.Key)
		}
	}

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
	if len(fields) >= 3 {
		fields[2].Input.SetValue(cfg.ExecutionMonitorEnabled.Value)
		fields[2].Value = fields[2].Input.Value()
	}
	if len(fields) >= 4 {
		fields[3].Input.SetValue(cfg.ExecutionMonitorSameToolLimit.Value)
		fields[3].Value = fields[3].Input.Value()
	}
	if len(fields) >= 5 {
		fields[4].Input.SetValue(cfg.ExecutionMonitorTotalToolLimit.Value)
		fields[4].Value = fields[4].Input.Value()
	}
	if len(fields) >= 6 {
		fields[5].Input.SetValue(cfg.MaxGeneralAgentToolCalls.Value)
		fields[5].Value = fields[5].Input.Value()
	}
	if len(fields) >= 7 {
		fields[6].Input.SetValue(cfg.MaxLimitedAgentToolCalls.Value)
		fields[6].Value = fields[6].Input.Value()
	}
	if len(fields) >= 8 {
		fields[7].Input.SetValue(cfg.AgentPlanningStepEnabled.Value)
		fields[7].Value = fields[7].Input.Value()
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
