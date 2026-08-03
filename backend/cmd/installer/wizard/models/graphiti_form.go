package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/go-units"
)

const (
	GraphitiURLPlaceholder       = "http://graphiti:8000"
	GraphitiTimeoutPlaceholder   = "30"
	GraphitiNeo4jUserPlaceholder = "neo4j"

	graphitiSemaphoreMin = 1
	graphitiSemaphoreMax = 10000
	graphitiWorkersMin   = 1
	graphitiWorkersMax   = 1024
	graphitiQueueMin     = 0
	graphitiQueueMax     = 2147483647
)

var (
	graphitiProviderPresets     = []string{"openai", "gemini", "custom", "litellm"}
	graphitiLogLevels           = []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	graphitiSearchScopes        = []string{"flowid", "all"}
	graphitiPolicyFields        = []string{"name", "source_description", "both"}
	graphitiPolicyActions       = []string{"REJECT", "SKIP_LLM", "PROCESS"}
	graphitiTaxonomySuggestions = []string{
		"STRUCTURAL,EVIDENCE,PROGRESS,ATTEMPT",
		"minimal",
		"full",
	}
)

// GraphitiFormModel represents the Graphiti configuration form
type GraphitiFormModel struct {
	*BaseScreen

	// screen-specific components
	deploymentList     list.Model
	deploymentDelegate *BaseListDelegate
}

// NewGraphitiFormModel creates a new Graphiti form model
func NewGraphitiFormModel(c controller.Controller, s styles.Styles, w window.Window) *GraphitiFormModel {
	m := &GraphitiFormModel{}

	m.BaseScreen = NewBaseScreen(c, s, w, m, m)
	m.initializeDeploymentList(s)

	return m
}

// initializeDeploymentList sets up the deployment type selection list
func (m *GraphitiFormModel) initializeDeploymentList(styles styles.Styles) {
	options := []BaseListOption{
		{Value: "embedded", Display: locale.MonitoringGraphitiEmbedded},
		{Value: "external", Display: locale.MonitoringGraphitiExternal},
		{Value: "disabled", Display: locale.MonitoringGraphitiDisabled},
	}

	m.deploymentDelegate = NewBaseListDelegate(
		styles.FormLabel.Align(lipgloss.Center),
		MinMenuWidth-6,
	)

	m.deploymentList = m.GetListHelper().CreateList(options, m.deploymentDelegate, MinMenuWidth-6, 3)

	config := m.GetController().GetGraphitiConfig()

	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)
}

// getSelectedDeploymentType returns the currently selected deployment type using the helper
func (m *GraphitiFormModel) getSelectedDeploymentType() string {
	selectedValue := m.GetListHelper().GetSelectedValue(&m.deploymentList)
	if selectedValue == "" {
		return "disabled"
	}

	return selectedValue
}

// BaseScreenHandler interface implementation

func (m *GraphitiFormModel) BuildForm() tea.Cmd {
	config := m.GetController().GetGraphitiConfig()
	fields := []FormField{}
	deploymentType := m.getSelectedDeploymentType()

	switch deploymentType {
	case "embedded":
		fields = append(fields, m.createIntegerField(config, "timeout",
			locale.MonitoringGraphitiTimeout, locale.MonitoringGraphitiTimeoutDesc, false, GraphitiTimeoutPlaceholder,
			1, 3600,
		))
		fields = append(fields, m.createSelectField(config, "llm_client_type",
			locale.MonitoringGraphitiLLMClientType, locale.MonitoringGraphitiLLMClientTypeDesc, graphitiProviderPresets,
		))
		fields = append(fields, m.createBooleanField(config, "separate_embedding",
			locale.MonitoringGraphitiSeparateEmbedding, locale.MonitoringGraphitiSeparateEmbeddingDesc,
		))
		fields = append(fields, m.createIntegerField(config, "semaphore_limit",
			locale.MonitoringGraphitiSemaphoreLimit, locale.MonitoringGraphitiSemaphoreLimitDesc, false, "",
			graphitiSemaphoreMin, graphitiSemaphoreMax,
		))
		fields = append(fields, m.createIntegerField(config, "ingest_worker_count",
			locale.MonitoringGraphitiIngestWorkerCount, locale.MonitoringGraphitiIngestWorkerCountDesc, false, "",
			graphitiWorkersMin, graphitiWorkersMax,
		))
		fields = append(fields, m.createSelectField(config, "log_level",
			locale.MonitoringGraphitiLogLevel, locale.MonitoringGraphitiLogLevelDesc, graphitiLogLevels,
		))
		fields = append(fields, m.createSelectField(config, "search_scope",
			locale.MonitoringGraphitiSearchScope, locale.MonitoringGraphitiSearchScopeDesc, graphitiSearchScopes,
		))
		fields = append(fields, m.createTextField(config, "ingest_policy_rules",
			locale.MonitoringGraphitiIngestPolicyRules, locale.MonitoringGraphitiIngestPolicyRulesDesc, false, "",
		))
		fields = append(fields, m.createSelectField(config, "ingest_policy_field",
			locale.MonitoringGraphitiIngestPolicyField, locale.MonitoringGraphitiIngestPolicyFieldDesc, graphitiPolicyFields,
		))
		fields = append(fields, m.createSelectField(config, "ingest_policy_default_action",
			locale.MonitoringGraphitiIngestDefaultAction, locale.MonitoringGraphitiIngestDefaultActionDesc, graphitiPolicyActions,
		))
		fields = append(fields, m.createIntegerField(config, "ingest_queue_max_size",
			locale.MonitoringGraphitiIngestQueueMaxSize, locale.MonitoringGraphitiIngestQueueMaxSizeDesc, false, "",
			graphitiQueueMin, graphitiQueueMax,
		))
		fields = append(fields, m.createSelectField(config, "taxonomy_layer_profile",
			locale.MonitoringGraphitiTaxonomyLayerProfile, locale.MonitoringGraphitiTaxonomyLayerProfileDesc,
			graphitiTaxonomySuggestions,
		))
		fields = append(fields, m.createTextField(config, "graphiti_cpus",
			locale.MonitoringGraphitiCPUs, locale.MonitoringGraphitiCPUsDesc, false, "",
		))
		fields = append(fields, m.createTextField(config, "graphiti_memory",
			locale.MonitoringGraphitiMemory, locale.MonitoringGraphitiMemoryDesc, false, "",
		))
		fields = append(fields, m.createTextField(config, "neo4j_user",
			locale.MonitoringGraphitiNeo4jUser, locale.MonitoringGraphitiNeo4jUserDesc, false, GraphitiNeo4jUserPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "neo4j_password",
			locale.MonitoringGraphitiNeo4jPassword, locale.MonitoringGraphitiNeo4jPasswordDesc, true, "",
		))
		fields = append(fields, m.createTextField(config, "neo4j_database",
			locale.MonitoringGraphitiNeo4jDatabase, locale.MonitoringGraphitiNeo4jDatabaseDesc, false, GraphitiNeo4jUserPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "neo4j_cpus",
			locale.MonitoringGraphitiNeo4jCPUs, locale.MonitoringGraphitiNeo4jCPUsDesc, false, "",
		))
		fields = append(fields, m.createTextField(config, "neo4j_memory",
			locale.MonitoringGraphitiNeo4jMemory, locale.MonitoringGraphitiNeo4jMemoryDesc, false, "",
		))
		fields = append(fields, m.createTextField(config, "neo4j_shm_size",
			locale.MonitoringGraphitiNeo4jShmSize, locale.MonitoringGraphitiNeo4jShmSizeDesc, false, "",
		))

	case "external":
		// External mode - requires connection details only
		fields = append(fields, m.createTextField(config, "url",
			locale.MonitoringGraphitiURL, locale.MonitoringGraphitiURLDesc, false, GraphitiURLPlaceholder,
		))
		fields = append(fields, m.createTextField(config, "timeout",
			locale.MonitoringGraphitiTimeout, locale.MonitoringGraphitiTimeoutDesc, false, GraphitiTimeoutPlaceholder,
		))

	case "disabled":
		// Disabled mode has no additional fields
	}

	m.SetFormFields(fields)
	return nil
}

func (m *GraphitiFormModel) createTextField(
	config *controller.GraphitiConfig, key, title, description string, masked bool, placeholder string,
) FormField {
	envVar := m.graphitiEnvVar(config, key)
	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)
	if placeholder != "" {
		input.Placeholder = placeholder
	}

	return FormField{
		Key:         key,
		Title:       title,
		Description: description,
		Required:    false,
		Masked:      masked,
		Input:       input,
		Value:       input.Value(),
	}
}

func (m *GraphitiFormModel) createBooleanField(
	config *controller.GraphitiConfig, key, title, description string,
) FormField {
	envVar := m.graphitiEnvVar(config, key)
	input := NewBooleanInput(m.GetStyles(), m.GetWindow(), envVar)
	return FormField{
		Key: key, Title: title, Description: description, Input: input,
		Value: input.Value(), Suggestions: input.AvailableSuggestions(),
	}
}

func (m *GraphitiFormModel) createSelectField(
	config *controller.GraphitiConfig, key, title, description string, suggestions []string,
) FormField {
	envVar := m.graphitiEnvVar(config, key)
	input := NewTextInput(m.GetStyles(), m.GetWindow(), envVar)
	input.ShowSuggestions = true
	input.SetSuggestions(suggestions)
	return FormField{
		Key: key, Title: title, Description: description, Input: input,
		Value: input.Value(), Suggestions: suggestions,
	}
}

func (m *GraphitiFormModel) createIntegerField(
	config *controller.GraphitiConfig,
	key, title, description string,
	masked bool,
	placeholder string,
	minValue, maxValue int,
) FormField {
	field := m.createTextField(config, key, title, description, masked, placeholder)
	envVar := m.graphitiEnvVar(config, key)
	if envVar.Default != "" {
		field.Input.Placeholder = fmt.Sprintf("%s (%d-%d)", envVar.Default, minValue, maxValue)
	} else {
		field.Input.Placeholder = fmt.Sprintf("(%d-%d)", minValue, maxValue)
	}
	return field
}

func (m *GraphitiFormModel) graphitiEnvVar(
	config *controller.GraphitiConfig, key string,
) loader.EnvVar {
	switch key {
	case "url":
		return config.GraphitiURL
	case "timeout":
		return config.Timeout
	case "llm_client_type":
		return config.LLMClientType
	case "separate_embedding":
		return config.SeparateEmbedding
	case "semaphore_limit":
		return config.SemaphoreLimit
	case "log_level":
		return config.LogLevel
	case "search_scope":
		return config.SearchScope
	case "ingest_policy_rules":
		return config.IngestPolicyRules
	case "ingest_policy_field":
		return config.IngestPolicyField
	case "ingest_policy_default_action":
		return config.IngestPolicyDefaultAction
	case "ingest_worker_count":
		return config.IngestWorkerCount
	case "ingest_queue_max_size":
		return config.IngestQueueMaxSize
	case "taxonomy_layer_profile":
		return config.TaxonomyLayerProfile
	case "graphiti_cpus":
		return config.GraphitiCPUs
	case "graphiti_memory":
		return config.GraphitiMemory
	case "neo4j_user":
		return config.Neo4jUser
	case "neo4j_password":
		return config.Neo4jPassword
	case "neo4j_database":
		return config.Neo4jDatabase
	case "neo4j_cpus":
		return config.Neo4jCPUs
	case "neo4j_memory":
		return config.Neo4jMemory
	case "neo4j_shm_size":
		return config.Neo4jShmSize
	default:
		return loader.EnvVar{}
	}
}

func (m *GraphitiFormModel) GetFormTitle() string {
	return locale.MonitoringGraphitiFormTitle
}

func (m *GraphitiFormModel) GetFormDescription() string {
	return locale.MonitoringGraphitiFormDescription
}

func (m *GraphitiFormModel) GetFormName() string {
	return locale.MonitoringGraphitiFormName
}

func (m *GraphitiFormModel) GetFormSummary() string {
	return ""
}

func (m *GraphitiFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringGraphitiFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.MonitoringGraphitiFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.MonitoringGraphitiFormOverview))

	return strings.Join(sections, "\n")
}

func (m *GraphitiFormModel) GetCurrentConfiguration() string {
	sections := []string{m.GetStyles().Subtitle.Render(m.GetFormName())}
	config := m.GetController().GetGraphitiConfig()

	appendValue := func(title string, envVar loader.EnvVar, masked bool) {
		value := envVar.Value
		style := m.GetStyles().Info
		if value == "" {
			value = envVar.Default
			style = m.GetStyles().Muted
		}
		if value == "" {
			return
		}
		if masked {
			value = strings.Repeat("*", min(len(value), 15))
			if len(envVar.Value) > 15 {
				value += "..."
			}
		}
		sections = append(sections, fmt.Sprintf("• %s: %s", title, style.Render(value)))
	}

	switch config.DeploymentType {
	case "embedded":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.MonitoringGraphitiEmbedded))
		appendValue(locale.MonitoringGraphitiTimeout, config.Timeout, false)
		appendValue(locale.MonitoringGraphitiLLMClientType, config.LLMClientType, false)
		appendValue(locale.MonitoringGraphitiSeparateEmbedding, config.SeparateEmbedding, false)
		appendValue(locale.MonitoringGraphitiSemaphoreLimit, config.SemaphoreLimit, false)
		appendValue(locale.MonitoringGraphitiIngestWorkerCount, config.IngestWorkerCount, false)
		appendValue(locale.MonitoringGraphitiSearchScope, config.SearchScope, false)
		appendValue(locale.MonitoringGraphitiIngestDefaultAction, config.IngestPolicyDefaultAction, false)
		appendValue(locale.MonitoringGraphitiIngestQueueMaxSize, config.IngestQueueMaxSize, false)
		appendValue(locale.MonitoringGraphitiTaxonomyLayerProfile, config.TaxonomyLayerProfile, false)
		appendValue(locale.MonitoringGraphitiCPUs, config.GraphitiCPUs, false)
		appendValue(locale.MonitoringGraphitiMemory, config.GraphitiMemory, false)
		appendValue(locale.MonitoringGraphitiNeo4jUser, config.Neo4jUser, false)
		appendValue(locale.MonitoringGraphitiNeo4jPassword, config.Neo4jPassword, true)
		appendValue(locale.MonitoringGraphitiNeo4jDatabase, config.Neo4jDatabase, false)
		appendValue(locale.MonitoringGraphitiNeo4jCPUs, config.Neo4jCPUs, false)
		appendValue(locale.MonitoringGraphitiNeo4jMemory, config.Neo4jMemory, false)
		appendValue(locale.MonitoringGraphitiNeo4jShmSize, config.Neo4jShmSize, false)

	case "external":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Success.Render(locale.MonitoringGraphitiExternal))
		appendValue(locale.MonitoringGraphitiURL, config.GraphitiURL, false)
		appendValue(locale.MonitoringGraphitiTimeout, config.Timeout, false)

	case "disabled":
		sections = append(sections, "• "+locale.UIMode+m.GetStyles().Warning.Render(locale.MonitoringGraphitiDisabled))
	}

	return strings.Join(sections, "\n")
}

func (m *GraphitiFormModel) IsConfigured() bool {
	config := m.GetController().GetGraphitiConfig()
	return config.DeploymentType != "disabled"
}

func (m *GraphitiFormModel) GetHelpContent() string {
	var sections []string
	deploymentType := m.getSelectedDeploymentType()

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MonitoringGraphitiFormTitle))
	sections = append(sections, "")
	sections = append(sections, locale.MonitoringGraphitiModeGuide)
	sections = append(sections, "")

	switch deploymentType {
	case "embedded":
		sections = append(sections, locale.MonitoringGraphitiEmbeddedHelp)
	case "external":
		sections = append(sections, locale.MonitoringGraphitiExternalHelp)
	case "disabled":
		sections = append(sections, locale.MonitoringGraphitiDisabledHelp)
	}

	return strings.Join(sections, "\n")
}

func (m *GraphitiFormModel) HandleSave() error {
	config := m.GetController().GetGraphitiConfig()
	deploymentType := m.getSelectedDeploymentType()
	fields := m.GetFormFields()

	newConfig := &controller.GraphitiConfig{
		DeploymentType: deploymentType,
		GraphitiURL:    config.GraphitiURL, Timeout: config.Timeout,
		LLMClientType: config.LLMClientType, SeparateEmbedding: config.SeparateEmbedding,
		SemaphoreLimit: config.SemaphoreLimit, LogLevel: config.LogLevel,
		SearchScope: config.SearchScope, IngestPolicyRules: config.IngestPolicyRules,
		IngestPolicyField:         config.IngestPolicyField,
		IngestPolicyDefaultAction: config.IngestPolicyDefaultAction,
		IngestWorkerCount:         config.IngestWorkerCount, IngestQueueMaxSize: config.IngestQueueMaxSize,
		TaxonomyLayerProfile: config.TaxonomyLayerProfile,
		GraphitiCPUs:         config.GraphitiCPUs, GraphitiMemory: config.GraphitiMemory,
		Neo4jUser: config.Neo4jUser, Neo4jPassword: config.Neo4jPassword,
		Neo4jDatabase: config.Neo4jDatabase, Neo4jURI: config.Neo4jURI,
		Neo4jCPUs: config.Neo4jCPUs, Neo4jMemory: config.Neo4jMemory,
		Neo4jShmSize: config.Neo4jShmSize, Installed: config.Installed,
	}

	for _, field := range fields {
		value := strings.TrimSpace(field.Input.Value())
		envVar := m.graphitiEnvVar(config, field.Key)
		if value == "" {
			value = envVar.Default
		}

		switch field.Key {
		case "url":
			if err := validateGraphitiURL(value); err != nil {
				return err
			}
			newConfig.GraphitiURL.Value = value
		case "timeout":
			if err := validateGraphitiInteger(value, locale.MonitoringGraphitiTimeout, 1, 3600); err != nil {
				return err
			}
			newConfig.Timeout.Value = value
		case "llm_client_type":
			if err := validateGraphitiEnum(value, locale.MonitoringGraphitiLLMClientType, graphitiProviderPresets); err != nil {
				return err
			}
			newConfig.LLMClientType.Value = value
		case "separate_embedding":
			if value != "true" && value != "false" {
				return fmt.Errorf("%s must be true or false", locale.MonitoringGraphitiSeparateEmbedding)
			}
			newConfig.SeparateEmbedding.Value = value
		case "semaphore_limit":
			if err := validateGraphitiInteger(value, locale.MonitoringGraphitiSemaphoreLimit, graphitiSemaphoreMin, graphitiSemaphoreMax); err != nil {
				return err
			}
			newConfig.SemaphoreLimit.Value = value
		case "log_level":
			value = strings.ToUpper(value)
			if err := validateGraphitiEnum(value, locale.MonitoringGraphitiLogLevel, graphitiLogLevels); err != nil {
				return err
			}
			newConfig.LogLevel.Value = value
		case "search_scope":
			value = strings.ToLower(value)
			if err := validateGraphitiEnum(value, locale.MonitoringGraphitiSearchScope, graphitiSearchScopes); err != nil {
				return err
			}
			newConfig.SearchScope.Value = value
		case "ingest_policy_rules":
			normalized, err := normalizeGraphitiPolicyRules(value)
			if err != nil {
				return err
			}
			newConfig.IngestPolicyRules.Value = normalized
		case "ingest_policy_field":
			value = strings.ToLower(value)
			if err := validateGraphitiEnum(value, locale.MonitoringGraphitiIngestPolicyField, graphitiPolicyFields); err != nil {
				return err
			}
			newConfig.IngestPolicyField.Value = value
		case "ingest_policy_default_action":
			value = strings.ToUpper(value)
			if err := validateGraphitiEnum(value, locale.MonitoringGraphitiIngestDefaultAction, graphitiPolicyActions); err != nil {
				return err
			}
			newConfig.IngestPolicyDefaultAction.Value = value
		case "ingest_worker_count":
			if err := validateGraphitiInteger(value, locale.MonitoringGraphitiIngestWorkerCount, graphitiWorkersMin, graphitiWorkersMax); err != nil {
				return err
			}
			newConfig.IngestWorkerCount.Value = value
		case "ingest_queue_max_size":
			if err := validateGraphitiInteger(value, locale.MonitoringGraphitiIngestQueueMaxSize, graphitiQueueMin, graphitiQueueMax); err != nil {
				return err
			}
			newConfig.IngestQueueMaxSize.Value = value
		case "taxonomy_layer_profile":
			normalized, err := normalizeGraphitiTaxonomyProfile(value)
			if err != nil {
				return err
			}
			newConfig.TaxonomyLayerProfile.Value = normalized
		case "graphiti_cpus":
			if err := validateGraphitiCPUs(value, locale.MonitoringGraphitiCPUs); err != nil {
				return err
			}
			newConfig.GraphitiCPUs.Value = value
		case "graphiti_memory":
			if err := validateGraphitiMemory(value, locale.MonitoringGraphitiMemory); err != nil {
				return err
			}
			newConfig.GraphitiMemory.Value = value
		case "neo4j_user":
			if value == "" {
				return fmt.Errorf("%s cannot be empty", locale.MonitoringGraphitiNeo4jUser)
			}
			newConfig.Neo4jUser.Value = value
		case "neo4j_password":
			if value == "" {
				return fmt.Errorf("%s cannot be empty", locale.MonitoringGraphitiNeo4jPassword)
			}
			newConfig.Neo4jPassword.Value = value
		case "neo4j_database":
			if value == "" {
				return fmt.Errorf("%s cannot be empty", locale.MonitoringGraphitiNeo4jDatabase)
			}
			newConfig.Neo4jDatabase.Value = value
		case "neo4j_cpus":
			if err := validateGraphitiCPUs(value, locale.MonitoringGraphitiNeo4jCPUs); err != nil {
				return err
			}
			newConfig.Neo4jCPUs.Value = value
		case "neo4j_memory":
			if err := validateGraphitiMemory(value, locale.MonitoringGraphitiNeo4jMemory); err != nil {
				return err
			}
			newConfig.Neo4jMemory.Value = value
		case "neo4j_shm_size":
			if err := validateGraphitiMemory(value, locale.MonitoringGraphitiNeo4jShmSize); err != nil {
				return err
			}
			newConfig.Neo4jShmSize.Value = value
		default:
			return fmt.Errorf("unknown Graphiti field: %s", field.Key)
		}
	}

	// save the configuration
	if err := m.GetController().UpdateGraphitiConfig(newConfig); err != nil {
		logger.Errorf("[GraphitiFormModel] SAVE: error updating graphiti config: %v", err)
		return err
	}

	logger.Log("[GraphitiFormModel] SAVE: success")
	return nil
}

func validateGraphitiURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be an HTTP or HTTPS URL", locale.MonitoringGraphitiURL)
	}
	return nil
}

func validateGraphitiEnum(value, fieldName string, allowed []string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
}

func validateGraphitiInteger(value, fieldName string, minValue, maxValue int) error {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minValue || parsed > maxValue {
		return fmt.Errorf("%s must be an integer between %d and %d", fieldName, minValue, maxValue)
	}
	return nil
}

func validateGraphitiCPUs(value, fieldName string) error {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return fmt.Errorf("%s must be a positive CPU value", fieldName)
	}
	return nil
}

func validateGraphitiMemory(value, fieldName string) error {
	bytes, err := units.RAMInBytes(value)
	if err != nil || bytes <= 0 {
		return fmt.Errorf("%s must be a positive memory value such as 2G or 512M", fieldName)
	}
	return nil
}

func normalizeGraphitiPolicyRules(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') ||
		(value[0] == '"' && value[len(value)-1] == '"')) {
		value = value[1 : len(value)-1]
	}

	var rules map[string]string
	if err := json.Unmarshal([]byte(value), &rules); err != nil {
		return "", fmt.Errorf("%s must be a JSON object: %w", locale.MonitoringGraphitiIngestPolicyRules, err)
	}
	for pattern, action := range rules {
		if strings.TrimSpace(pattern) == "" {
			return "", fmt.Errorf("%s cannot contain an empty pattern", locale.MonitoringGraphitiIngestPolicyRules)
		}
		action = strings.ToUpper(strings.TrimSpace(action))
		if err := validateGraphitiEnum(action, locale.MonitoringGraphitiIngestPolicyRules, graphitiPolicyActions); err != nil {
			return "", err
		}
		rules[pattern] = action
	}
	normalized, err := json.Marshal(rules)
	if err != nil {
		return "", fmt.Errorf("failed to normalize %s: %w", locale.MonitoringGraphitiIngestPolicyRules, err)
	}
	return string(normalized), nil
}

func normalizeGraphitiTaxonomyProfile(value string) (string, error) {
	value = strings.TrimSpace(value)
	alias := strings.ToLower(value)
	if alias == "full" || alias == "all" || alias == "minimal" {
		return alias, nil
	}

	validClasses := map[string]bool{
		"STRUCTURAL":   true,
		"EVIDENCE":     true,
		"PROGRESS":     true,
		"AGENT_ACTION": true,
		"META":         true,
		"ATTEMPT":      true,
		"NARRATIVE":    true,
	}
	parts := strings.Split(value, ",")
	normalized := make([]string, 0, len(parts))
	seen := make(map[string]bool)
	for _, part := range parts {
		className := strings.ToUpper(strings.TrimSpace(part))
		if !validClasses[className] {
			return "", fmt.Errorf("%s contains unknown class %q", locale.MonitoringGraphitiTaxonomyLayerProfile, part)
		}
		if !seen[className] {
			normalized = append(normalized, className)
			seen[className] = true
		}
	}
	if len(normalized) == 0 {
		return "", fmt.Errorf("%s cannot be empty", locale.MonitoringGraphitiTaxonomyLayerProfile)
	}
	return strings.Join(normalized, ","), nil
}

func (m *GraphitiFormModel) HandleReset() {
	// reset config to defaults
	config := m.GetController().ResetGraphitiConfig()

	// reset deployment selection
	m.GetListHelper().SelectByValue(&m.deploymentList, config.DeploymentType)

	// rebuild form with reset deployment type
	m.BuildForm()
}

func (m *GraphitiFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// additional validation could be added here if needed
}

func (m *GraphitiFormModel) GetFormFields() []FormField {
	return m.BaseScreen.fields
}

func (m *GraphitiFormModel) SetFormFields(fields []FormField) {
	m.BaseScreen.fields = fields
}

// BaseListHandler interface implementation

func (m *GraphitiFormModel) GetList() *list.Model {
	return &m.deploymentList
}

func (m *GraphitiFormModel) GetListDelegate() *BaseListDelegate {
	return m.deploymentDelegate
}

func (m *GraphitiFormModel) OnListSelectionChanged(oldSelection, newSelection string) {
	// rebuild form when deployment type changes
	m.BuildForm()
}

func (m *GraphitiFormModel) GetListTitle() string {
	return locale.MonitoringGraphitiDeploymentType
}

func (m *GraphitiFormModel) GetListDescription() string {
	return locale.MonitoringGraphitiDeploymentTypeDesc
}

// Update method - handle screen-specific input
func (m *GraphitiFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// handle list input first (if focused on list)
		if cmd := m.HandleListInput(msg); cmd != nil {
			return m, cmd
		}

		// then handle field input
		if cmd := m.HandleFieldInput(msg); cmd != nil {
			return m, cmd
		}
	}

	// delegate to base screen for common handling
	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// Compile-time interface validation
var _ BaseScreenModel = (*GraphitiFormModel)(nil)
var _ BaseScreenHandler = (*GraphitiFormModel)(nil)
var _ BaseListHandler = (*GraphitiFormModel)(nil)
