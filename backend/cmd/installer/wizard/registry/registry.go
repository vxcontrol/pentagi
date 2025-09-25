package registry

import (
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/models"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

type Registry interface {
	models.Registry
	HandleMsg(msg tea.Msg) tea.Cmd
}

type registry struct {
	files      files.Files
	styles     styles.Styles
	window     window.Window
	processor  processor.ProcessorModel
	controller controller.Controller
	screens    map[models.ScreenID]models.BaseScreenModel
}

func NewRegistry(
	c controller.Controller, s styles.Styles, w window.Window, f files.Files, p processor.ProcessorModel,
) Registry {
	r := &registry{
		files:      f,
		styles:     s,
		window:     w,
		processor:  p,
		controller: c,
		screens:    make(map[models.ScreenID]models.BaseScreenModel),
	}

	r.initScreens()

	return r
}

func (r *registry) initScreens() {
	// Core Screens
	r.screens[models.WelcomeScreen] = models.NewWelcomeModel(r.controller, r.styles, r.window)
	r.screens[models.EULAScreen] = models.NewEULAModel(r.controller, r.styles, r.window, r.files)
	r.screens[models.MainMenuScreen] = models.NewMainMenuModel(r.controller, r.styles, r.window, r)

	// LLM Provider Forms
	r.screens[models.LLMProvidersScreen] = models.NewLLMProvidersModel(r.controller, r.styles, r.window, r)
	r.screens[models.LLMProviderOpenAIScreen] = models.NewLLMProviderFormModel(r.controller, r.styles, r.window, models.LLMProviderOpenAI)
	r.screens[models.LLMProviderAnthropicScreen] = models.NewLLMProviderFormModel(r.controller, r.styles, r.window, models.LLMProviderAnthropic)
	r.screens[models.LLMProviderGeminiScreen] = models.NewLLMProviderFormModel(r.controller, r.styles, r.window, models.LLMProviderGemini)
	r.screens[models.LLMProviderBedrockScreen] = models.NewLLMProviderFormModel(r.controller, r.styles, r.window, models.LLMProviderBedrock)
	r.screens[models.LLMProviderOllamaScreen] = models.NewLLMProviderFormModel(r.controller, r.styles, r.window, models.LLMProviderOllama)
	r.screens[models.LLMProviderCustomScreen] = models.NewLLMProviderFormModel(r.controller, r.styles, r.window, models.LLMProviderCustom)

	// Summarizer Forms
	r.screens[models.SummarizerScreen] = models.NewSummarizerModel(r.controller, r.styles, r.window, r)
	r.screens[models.SummarizerGeneralScreen] = models.NewSummarizerFormModel(r.controller, r.styles, r.window, controller.SummarizerTypeGeneral)
	r.screens[models.SummarizerAssistantScreen] = models.NewSummarizerFormModel(r.controller, r.styles, r.window, controller.SummarizerTypeAssistant)

	// Monitoring Forms
	r.screens[models.MonitoringScreen] = models.NewMonitoringModel(r.controller, r.styles, r.window, r)
	r.screens[models.LangfuseScreen] = models.NewLangfuseFormModel(r.controller, r.styles, r.window)
	r.screens[models.ObservabilityScreen] = models.NewObservabilityFormModel(r.controller, r.styles, r.window)

	// Tools Forms
	r.screens[models.ToolsScreen] = models.NewToolsModel(r.controller, r.styles, r.window, r)
	r.screens[models.AIAgentsSettingsFormScreen] = models.NewAIAgentsSettingsFormModel(r.controller, r.styles, r.window)
	r.screens[models.SearchEnginesFormScreen] = models.NewSearchEnginesFormModel(r.controller, r.styles, r.window)
	r.screens[models.ScraperFormScreen] = models.NewScraperFormModel(r.controller, r.styles, r.window)
	r.screens[models.DockerFormScreen] = models.NewDockerFormModel(r.controller, r.styles, r.window)

	// Embedder Form
	r.screens[models.EmbedderFormScreen] = models.NewEmbedderFormModel(r.controller, r.styles, r.window)

	// Server Settings Form
	r.screens[models.ServerSettingsScreen] = models.NewServerSettingsFormModel(r.controller, r.styles, r.window)

	// Changes Form
	r.screens[models.ApplyChangesScreen] = models.NewApplyChangesFormModel(r.controller, r.styles, r.window, r.processor)

	// Maintenance
	r.screens[models.MaintenanceScreen] = models.NewMaintenanceModel(r.controller, r.styles, r.window, r)
	r.screens[models.ResetPasswordScreen] = models.NewResetPasswordModel(r.controller, r.styles, r.window, r.processor)

	// Processor Operation Forms
	processorOperationForms := []models.ScreenID{
		models.InstallPentagiScreen,
		models.StartPentagiScreen,
		models.StopPentagiScreen,
		models.RestartPentagiScreen,
		models.DownloadWorkerImageScreen,
		models.UpdateWorkerImageScreen,
		models.UpdatePentagiScreen,
		models.UpdateInstallerScreen,
		models.FactoryResetScreen,
		models.RemovePentagiScreen,
		models.PurgePentagiScreen,
	}
	for _, id := range processorOperationForms {
		r.screens[id] = r.initProcessorOperationForm(id)
	}
}

func (r *registry) initProcessorOperationForm(id models.ScreenID) models.BaseScreenModel {
	// handle parameterized screens
	args := id.GetArgs()
	if len(args) < 2 {
		return r.initMockScreen()
	}

	stack := processor.ProductStack(args[0])
	operation := processor.ProcessorOperation(args[1])

	screen := models.NewProcessorOperationFormModel(r.controller, r.styles, r.window, r.processor, stack, operation)
	r.screens[id] = screen
	return screen
}

// initMockScreen initializes unknown screen with mock data
func (r *registry) initMockScreen() models.BaseScreenModel {
	title := locale.MockScreenTitle
	description := locale.MockScreenDescription

	return models.NewMockFormModel(r.controller, r.styles, r.window, title, title, description)
}

func (r *registry) GetScreen(id models.ScreenID) models.BaseScreenModel {
	if screen, ok := r.screens[id]; ok {
		return screen
	}

	screen := r.initMockScreen()
	r.screens[id] = screen

	return screen
}

// HandleMsg handles system messages only for all screens in the registry
func (r *registry) HandleMsg(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for _, screen := range r.screens {
		_, cmd := screen.Update(msg) // ignore updated model, save previous state
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) != 0 {
		return tea.Batch(cmds...)
	}

	return nil
}
