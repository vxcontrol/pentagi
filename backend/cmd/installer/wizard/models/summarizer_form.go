package models

import (
	"fmt"
	"strconv"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/models/helpers"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
	"pentagi/pkg/csum"

	tea "github.com/charmbracelet/bubbletea"
)

// SummarizerFormModel represents the Summarizer configuration form
type SummarizerFormModel struct {
	*BaseScreen

	// screen-specific components
	summarizerType controller.SummarizerType
	typeName       string
}

// NewSummarizerFormModel creates a new Summarizer form model
func NewSummarizerFormModel(
	c controller.Controller, s styles.Styles, w window.Window, st controller.SummarizerType,
) *SummarizerFormModel {
	tn := locale.SummarizerTypeGeneralName
	if st == controller.SummarizerTypeAssistant {
		tn = locale.SummarizerTypeAssistantName
	}

	m := &SummarizerFormModel{
		summarizerType: st,
		typeName:       tn,
	}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// Helper functions for working with loader.EnvVar

func (m *SummarizerFormModel) envVarToBool(envVar loader.EnvVar) bool {
	if envVar.Value != "" {
		return envVar.Value == "true"
	}
	return envVar.Default == "true"
}

func (m *SummarizerFormModel) envVarToInt(envVar loader.EnvVar) int {
	if envVar.Value != "" {
		if val, err := strconv.Atoi(envVar.Value); err == nil {
			return val
		}
	}
	if envVar.Default != "" {
		if val, err := strconv.Atoi(envVar.Default); err == nil {
			return val
		}
	}
	return 0
}

func (m *SummarizerFormModel) formatBytes(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * 1024
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func (m *SummarizerFormModel) formatBytesFromEnvVar(envVar loader.EnvVar) string {
	return m.formatBytes(m.envVarToInt(envVar))
}

func (m *SummarizerFormModel) formatBooleanStatus(value bool) string {
	if value {
		return m.GetStyles().Success.Render(locale.StatusEnabled)
	}
	return m.GetStyles().Warning.Render(locale.StatusDisabled)
}

func (m *SummarizerFormModel) formatNumber(num int) string {
	if num >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	return fmt.Sprintf("%d", num)
}

// BaseScreenHandler interface implementation

func (m *SummarizerFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetSummarizerConfig(m.summarizerType)
	fields := []FormField{}

	// Preserve Last Section (common for both types)
	fields = append(fields, m.createBooleanField("preserve_last",
		locale.SummarizerFormPreserveLast,
		locale.SummarizerFormPreserveLastDesc,
		config.PreserveLast,
	))

	// General-specific fields
	if m.summarizerType == controller.SummarizerTypeGeneral {
		// Use QA Pairs
		fields = append(fields, m.createBooleanField("use_qa",
			locale.SummarizerFormUseQA,
			locale.SummarizerFormUseQADesc,
			config.UseQA,
		))

		// Summarize Human in QA
		fields = append(fields, m.createBooleanField("sum_human_in_qa",
			locale.SummarizerFormSumHumanInQA,
			locale.SummarizerFormSumHumanInQADesc,
			config.SumHumanInQA,
		))
	}

	// Size settings
	fields = append(fields, m.createIntegerField("last_sec_bytes",
		locale.SummarizerFormLastSecBytes,
		locale.SummarizerFormLastSecBytesDesc,
		config.LastSecBytes,
		1024,    // min: 1KB
		1048576, // max: 1MB
	))

	fields = append(fields, m.createIntegerField("max_bp_bytes",
		locale.SummarizerFormMaxBPBytes,
		locale.SummarizerFormMaxBPBytesDesc,
		config.MaxBPBytes,
		512,     // min: 512B
		1048576, // max: 1MB
	))

	fields = append(fields, m.createIntegerField("max_qa_bytes",
		locale.SummarizerFormMaxQABytes,
		locale.SummarizerFormMaxQABytesDesc,
		config.MaxQABytes,
		1024,   // min: 1KB
		524288, // max: 512KB
	))

	// Count settings
	fields = append(fields, m.createIntegerField("max_qa_sections",
		locale.SummarizerFormMaxQASections,
		locale.SummarizerFormMaxQASectionsDesc,
		config.MaxQASections,
		1,  // min: 1
		50, // max: 50
	))

	fields = append(fields, m.createIntegerField("keep_qa_sections",
		locale.SummarizerFormKeepQASections,
		locale.SummarizerFormKeepQASectionsDesc,
		config.KeepQASections,
		1,  // min: 1
		20, // max: 20
	))

	m.SetFormFields(fields)
	return nil
}

func (m *SummarizerFormModel) createBooleanField(key, title, description string, envVar loader.EnvVar) FormField {
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

func (m *SummarizerFormModel) createIntegerField(key, title, description string, envVar loader.EnvVar, min, max int) FormField {
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

func (m *SummarizerFormModel) GetFormTitle() string {
	if m.summarizerType == controller.SummarizerTypeAssistant {
		return locale.SummarizerFormAssistantTitle
	}
	return locale.SummarizerFormGeneralTitle
}

func (m *SummarizerFormModel) GetFormDescription() string {
	return fmt.Sprintf(locale.SummarizerFormDescription, m.typeName)
}

func (m *SummarizerFormModel) GetFormName() string {
	return m.typeName
}

func (m *SummarizerFormModel) GetFormSummary() string {
	return m.calculateTokenEstimate()
}

func (m *SummarizerFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.typeName))
	sections = append(sections, "")

	if m.summarizerType == controller.SummarizerTypeAssistant {
		sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.SummarizerTypeAssistantDesc))
		sections = append(sections, "")
		sections = append(sections, m.styles.Paragraph.Render(locale.SummarizerTypeAssistantInfo))
	} else {
		sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.SummarizerTypeGeneralDesc))
		sections = append(sections, "")
		sections = append(sections, m.styles.Paragraph.Render(locale.SummarizerTypeGeneralInfo))
	}

	return strings.Join(sections, "\n")
}

func (m *SummarizerFormModel) GetCurrentConfiguration() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.GetFormName()))

	config := m.GetController().GetSummarizerConfig(m.summarizerType)

	sections = append(sections, fmt.Sprintf("• %s: %s",
		locale.SummarizerFormPreserveLast,
		m.GetStyles().Info.Render(m.formatBooleanStatus(m.envVarToBool(config.PreserveLast)))))

	if m.summarizerType == controller.SummarizerTypeGeneral {
		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.SummarizerFormUseQA,
			m.GetStyles().Info.Render(m.formatBooleanStatus(m.envVarToBool(config.UseQA)))))

		sections = append(sections, fmt.Sprintf("• %s: %s",
			locale.SummarizerFormSumHumanInQA,
			m.GetStyles().Info.Render(m.formatBooleanStatus(m.envVarToBool(config.SumHumanInQA)))))
	}

	sections = append(sections, fmt.Sprintf("• %s: %s",
		locale.SummarizerFormLastSecBytes,
		m.GetStyles().Info.Render(m.formatBytesFromEnvVar(config.LastSecBytes))))

	sections = append(sections, fmt.Sprintf("• %s: %s",
		locale.SummarizerFormMaxBPBytes,
		m.GetStyles().Info.Render(m.formatBytesFromEnvVar(config.MaxBPBytes))))

	sections = append(sections, fmt.Sprintf("• %s: %s",
		locale.SummarizerFormMaxQABytes,
		m.GetStyles().Info.Render(m.formatBytesFromEnvVar(config.MaxQABytes))))

	sections = append(sections, fmt.Sprintf("• %s: %s",
		locale.SummarizerFormMaxQASections,
		m.GetStyles().Info.Render(strconv.Itoa(m.envVarToInt(config.MaxQASections)))))

	sections = append(sections, fmt.Sprintf("• %s: %s",
		locale.SummarizerFormKeepQASections,
		m.GetStyles().Info.Render(strconv.Itoa(m.envVarToInt(config.KeepQASections)))))

	return strings.Join(sections, "\n")
}

func (m *SummarizerFormModel) IsConfigured() bool {
	// summarizer is always considered configured since it has defaults
	return true
}

func (m *SummarizerFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(fmt.Sprintf(locale.SummarizerFormDescription, m.typeName)))
	sections = append(sections, "")

	if m.summarizerType == controller.SummarizerTypeAssistant {
		sections = append(sections, locale.SummarizerFormAssistantHelp)
	} else {
		sections = append(sections, locale.SummarizerFormGeneralHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *SummarizerFormModel) HandleSave() error {
	config := m.GetController().GetSummarizerConfig(m.summarizerType)
	fields := m.GetFormFields()

	// create a working copy of the current config to modify
	newConfig := &controller.SummarizerConfig{
		Type: m.summarizerType,
		// copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
		PreserveLast:   config.PreserveLast,
		UseQA:          config.UseQA,
		SumHumanInQA:   config.SumHumanInQA,
		LastSecBytes:   config.LastSecBytes,
		MaxBPBytes:     config.MaxBPBytes,
		MaxQABytes:     config.MaxQABytes,
		MaxQASections:  config.MaxQASections,
		KeepQASections: config.KeepQASections,
	}

	// update field values based on form input
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())

		switch field.Key {
		case "preserve_last":
			if err := m.validateBooleanField(value, locale.SummarizerFormPreserveLast); err != nil {
				return err
			}
			newConfig.PreserveLast.Value = value

		case "use_qa":
			if err := m.validateBooleanField(value, locale.SummarizerFormUseQA); err != nil {
				return err
			}
			newConfig.UseQA.Value = value

		case "sum_human_in_qa":
			if err := m.validateBooleanField(value, locale.SummarizerFormSumHumanInQA); err != nil {
				return err
			}
			newConfig.SumHumanInQA.Value = value

		case "last_sec_bytes":
			if val, err := m.validateIntegerField(value, locale.SummarizerFormLastSecBytes, 1024, 1048576); err != nil {
				return err
			} else {
				newConfig.LastSecBytes.Value = strconv.Itoa(val)
			}

		case "max_bp_bytes":
			if val, err := m.validateIntegerField(value, locale.SummarizerFormMaxBPBytes, 512, 1048576); err != nil {
				return err
			} else {
				newConfig.MaxBPBytes.Value = strconv.Itoa(val)
			}

		case "max_qa_bytes":
			if val, err := m.validateIntegerField(value, locale.SummarizerFormMaxQABytes, 1024, 524288); err != nil {
				return err
			} else {
				newConfig.MaxQABytes.Value = strconv.Itoa(val)
			}

		case "max_qa_sections":
			if val, err := m.validateIntegerField(value, locale.SummarizerFormMaxQASections, 1, 50); err != nil {
				return err
			} else {
				newConfig.MaxQASections.Value = strconv.Itoa(val)
			}

		case "keep_qa_sections":
			if val, err := m.validateIntegerField(value, locale.SummarizerFormKeepQASections, 1, 20); err != nil {
				return err
			} else {
				newConfig.KeepQASections.Value = strconv.Itoa(val)
			}
		}
	}

	// save the configuration
	if err := m.GetController().UpdateSummarizerConfig(newConfig); err != nil {
		logger.Errorf("[SummarizerFormModel] SAVE: error updating summarizer config: %v", err)
		return err
	}

	logger.Log("[SummarizerFormModel] SAVE: success")
	return nil
}

func (m *SummarizerFormModel) validateBooleanField(value, fieldName string) error {
	if value != "" && value != "true" && value != "false" {
		return fmt.Errorf("invalid boolean value for %s: %s (must be 'true' or 'false')", fieldName, value)
	}
	return nil
}

func (m *SummarizerFormModel) validateIntegerField(value, fieldName string, min, max int) (int, error) {
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

func (m *SummarizerFormModel) HandleReset() {
	// reset config to defaults
	m.GetController().ResetSummarizerConfig(m.summarizerType)

	// rebuild form with reset values
	m.BuildForm()
}

func (m *SummarizerFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *SummarizerFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *SummarizerFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// Update method - handle screen-specific input
func (m *SummarizerFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// then handle field input
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}

	// delegate to base screen for common handling
	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// Helper methods

func (m *SummarizerFormModel) calculateTokenEstimate() string {
	fields := m.GetFormFields()
	if len(fields) == 0 {
		return ""
	}

	// Build configuration from current form values
	cscfg := m.buildConfigFromForm(fields)

	// For assistant type, force UseQA=true and SummHumanInQA=false
	if m.summarizerType == controller.SummarizerTypeAssistant {
		cscfg.UseQA = true
		cscfg.SummHumanInQA = false
	}

	// Calculate estimate using the helper
	estimate := helpers.CalculateContextEstimate(cscfg)

	// Format the estimate with localization
	var tokenRange string
	if estimate.MinTokens == estimate.MaxTokens {
		tokenRange = fmt.Sprintf(locale.SummarizerContextTokenRange,
			m.formatNumber(estimate.MinTokens),
		)
	} else {
		tokenRange = fmt.Sprintf(locale.SummarizerContextTokenRangeMinMax,
			m.formatNumber(estimate.MinTokens),
			m.formatNumber(estimate.MaxTokens),
		)
	}

	// Add context size guidance
	var guidance string
	if estimate.MaxTokens > 200_000 {
		guidance = locale.SummarizerContextRequires256K
	} else if estimate.MaxTokens > 120_000 {
		guidance = locale.SummarizerContextRequires128K
	} else if estimate.MaxTokens > 60_000 {
		guidance = locale.SummarizerContextRequires64K
	} else if estimate.MaxTokens > 30_000 {
		guidance = locale.SummarizerContextRequires32K
	} else if estimate.MaxTokens > 14_000 {
		guidance = locale.SummarizerContextRequires16K
	} else {
		guidance = locale.SummarizerContextFitsIn8K
	}

	return m.GetStyles().Info.Render(fmt.Sprintf(locale.SummarizerContextEstimatedSize, tokenRange, guidance))
}

func (m *SummarizerFormModel) buildConfigFromForm(fields []FormField) csum.SummarizerConfig {
	config := m.GetController().GetSummarizerConfig(m.summarizerType)

	// Start with current cscfg values as base
	cscfg := csum.SummarizerConfig{
		PreserveLast:   m.envVarToBool(config.PreserveLast),
		UseQA:          m.envVarToBool(config.UseQA),
		SummHumanInQA:  m.envVarToBool(config.SumHumanInQA),
		LastSecBytes:   m.envVarToInt(config.LastSecBytes),
		MaxBPBytes:     m.envVarToInt(config.MaxBPBytes),
		MaxQABytes:     m.envVarToInt(config.MaxQABytes),
		MaxQASections:  m.envVarToInt(config.MaxQASections),
		KeepQASections: m.envVarToInt(config.KeepQASections),
	}

	// Override with current form values where available
	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())
		if value == "" {
			continue // Keep config value if form field is empty
		}

		switch field.Key {
		case "preserve_last":
			cscfg.PreserveLast = (value == "true")

		case "use_qa":
			cscfg.UseQA = (value == "true")

		case "sum_human_in_qa":
			cscfg.SummHumanInQA = (value == "true")

		case "last_sec_bytes":
			if intVal, err := strconv.Atoi(value); err == nil && intVal > 0 {
				cscfg.LastSecBytes = intVal
			}

		case "max_bp_bytes":
			if intVal, err := strconv.Atoi(value); err == nil && intVal > 0 {
				cscfg.MaxBPBytes = intVal
			}

		case "max_qa_bytes":
			if intVal, err := strconv.Atoi(value); err == nil && intVal > 0 {
				cscfg.MaxQABytes = intVal
			}

		case "max_qa_sections":
			if intVal, err := strconv.Atoi(value); err == nil && intVal > 0 {
				cscfg.MaxQASections = intVal
			}

		case "keep_qa_sections":
			if intVal, err := strconv.Atoi(value); err == nil && intVal > 0 {
				cscfg.KeepQASections = intVal
			}
		}
	}

	return cscfg
}

// Compile-time interface validation
var _ BaseScreenModel = (*SummarizerFormModel)(nil)
var _ BaseScreenHandler = (*SummarizerFormModel)(nil)
