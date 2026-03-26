package wizard

import (
	"context"
	"fmt"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/navigator"
	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/models"
	"pentagi/cmd/installer/wizard/registry"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	BaseHeaderHeight = 2
	BaseFooterHeight = 1
	MinHeaderHeight  = 1
)

// App represents the main wizard application
type App struct {
	files        files.Files
	styles       styles.Styles
	window       window.Window
	registry     registry.Registry
	navigator    navigator.Navigator
	processor    processor.ProcessorModel
	controller   controller.Controller
	currentModel models.BaseScreenModel
	hotkeys      map[string]string
}

func NewApp(appState state.State, checkResult checker.CheckResult, files files.Files) *App {
	styles := styles.New()
	window := window.New()
	navigator := navigator.NewNavigator(appState, checkResult)
	controller := controller.NewController(appState, files, checkResult)
	processor := processor.NewProcessorModel(appState, controller.GetChecker(), files)
	registry := registry.NewRegistry(controller, styles, window, files, processor)

	if len(navigator.GetStack()) == 0 {
		navigator.Push(models.WelcomeScreen)
	}

	app := &App{
		files:      files,
		styles:     styles,
		window:     window,
		registry:   registry,
		navigator:  navigator,
		processor:  processor,
		controller: controller,
	}

	app.initHotkeysLocale()
	app.updateScreenMargins()

	app.currentModel = registry.GetScreen(navigator.Current())

	return app
}

func (app *App) initHotkeysLocale() {
	app.hotkeys = map[string]string{
		"up|down":     locale.NavUpDown,
		"left|right":  locale.NavLeftRight,
		"pgup|pgdown": locale.NavPgUpPgDown,
		"home|end":    locale.NavHomeEnd,
		"enter":       locale.NavEnter,
		"y|n":         locale.NavYn,
		"ctrl+c":      locale.NavCtrlC,
		"ctrl+s":      locale.NavCtrlS,
		"ctrl+r":      locale.NavCtrlR,
		"ctrl+h":      locale.NavCtrlH,
		"tab":         locale.NavTab,
	}
}

// updateScreenMargins calculates and sets header/footer margins based on current screen
func (app *App) updateScreenMargins() {
	app.window.SetHeaderHeight(lipgloss.Height(app.renderHeader()))
	app.window.SetFooterHeight(lipgloss.Height(app.renderFooter()))
}

func (app *App) Init() tea.Cmd {
	if cmd := app.currentModel.Init(); cmd != nil {
		return tea.Batch(cmd, tea.WindowSize())
	}

	return tea.WindowSize()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		app.window.SetWindowSize(msg.Width, msg.Height)

		// update content size
		msg.Width, msg.Height = app.window.GetContentSize()

		// update margins for current screen (header/footer might change with new size)
		app.updateScreenMargins()
		// forward the resize message to all screens
		return app, app.registry.HandleMsg(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			logger.Log("[App] QUIT")
			return app, tea.Quit
		case "esc":
			logger.Log("[App] ESC: %s", app.navigator.Current())
			if app.navigator.Current() != models.WelcomeScreen && app.navigator.CanGoBack() {
				// go back to previous screen
				targetScreen := app.navigator.Pop()
				app.currentModel = app.registry.GetScreen(targetScreen)
				logger.Log("[App] ESC: going back to %s", targetScreen)

				// update margins for the new screen
				app.updateScreenMargins()
				// soft initialize the new screen to synchronize with state
				return app, app.currentModel.Init()
			}
		}

	case models.NavigationMsg: // massages from screens
		if msg.GoBack && app.navigator.CanGoBack() {
			app.currentModel = app.registry.GetScreen(app.navigator.Pop())
		}
		if msg.Target != "" {
			app.navigator.Push(msg.Target)
			app.currentModel = app.registry.GetScreen(msg.Target)
		}

		// update margins for the new screen
		app.updateScreenMargins()
		// soft initialize the new screen to synchronize with state
		return app, app.currentModel.Init()
	}

	return app, app.forwardMsgToCurrentModel(msg)
}

func (app *App) View() string {
	if app.currentModel == nil {
		return locale.UILoading
	}

	// all screens have unified header/footer management
	header := app.renderHeader()
	footer := app.renderFooter()
	if !app.window.IsShowHeader() {
		header = ""
	}

	content := app.currentModel.View()
	contentArea := app.styles.Content.
		Width(app.window.GetContentWidth()).
		Height(app.window.GetContentHeight()).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}

func (app *App) forwardMsgToCurrentModel(msg tea.Msg) tea.Cmd {
	if app.currentModel != nil {
		model, cmd := app.currentModel.Update(msg)
		if newModel := models.RestoreModel(model); newModel != nil {
			app.currentModel = newModel
		}

		return cmd
	}

	return nil
}

func (app *App) renderHeader() string {
	currentScreen := app.navigator.Current()
	baseScreen := currentScreen.GetScreen()
	windowWidth := app.window.GetWindowWidth()

	switch models.ScreenID(baseScreen) {
	case models.WelcomeScreen:
		return app.styles.RenderASCIILogo(windowWidth)
	default:
		// other screens use text title
		return app.styles.Header.Width(windowWidth).Render(app.getScreenTitle())
	}
}

func (app *App) renderFooter() string {
	var actions []string
	var progressInfo string

	// add special progress info for EULA screen
	currentScreen := app.navigator.Current()
	if currentScreen.GetScreen() == string(models.EULAScreen) {
		if eulaModel, ok := app.currentModel.(*models.EULAModel); ok {
			_, atEnd, percent := eulaModel.GetScrollInfo()
			progressInfo = fmt.Sprintf(locale.EULAProgress, percent)
			if atEnd {
				progressInfo += locale.EULAProgressComplete
			}
			actions = append(actions, progressInfo)
		}
	}

	// add navigation actions
	if app.navigator.CanGoBack() && currentScreen.GetScreen() != string(models.WelcomeScreen) {
		actions = append(actions, locale.NavBack)
	}
	actions = append(actions, locale.NavExit)

	// get hotkeys from current screen model
	if app.currentModel != nil {
		hotkeys := app.currentModel.GetFormHotKeys()
		for _, hotkey := range hotkeys {
			if localeHotKey, ok := app.hotkeys[hotkey]; ok {
				actions = append(actions, localeHotKey)
			}
		}
	}

	return app.styles.RenderFooter(actions, app.window.GetWindowWidth())
}

func (app *App) getScreenTitle() string {
	if app.currentModel != nil {
		return app.currentModel.GetFormTitle()
	}
	return locale.WelcomeFormTitle
}

func Run(ctx context.Context, appState state.State, checkResult checker.CheckResult, files files.Files) error {
	app := NewApp(appState, checkResult, files)

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil { // ignore the return model, use app.currentModel instead
		return fmt.Errorf("failed to run installer wizard: %w", err)
	}

	return nil
}
