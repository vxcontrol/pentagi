package models

import (
	"context"
	"fmt"
	"strings"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ResetPasswordHandler handles the password reset functionality
type ResetPasswordHandler struct {
	controller controller.Controller
	styles     styles.Styles
	window     window.Window
	summary    string
	fields     []FormField
}

// NewResetPasswordHandler creates a new password reset handler
func NewResetPasswordHandler(c controller.Controller, s styles.Styles, w window.Window) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		controller: c,
		styles:     s,
		window:     w,
	}
}

// BuildForm creates form fields for password reset
func (h *ResetPasswordHandler) BuildForm() tea.Cmd {
	// create text input for new password
	newPasswordInput := NewTextInput(h.styles, h.window, loader.EnvVar{})
	newPasswordInput.EchoMode = textinput.EchoPassword
	newPasswordInput.EchoCharacter = '•'

	// create text input for confirm password
	confirmPasswordInput := NewTextInput(h.styles, h.window, loader.EnvVar{})
	confirmPasswordInput.EchoMode = textinput.EchoPassword
	confirmPasswordInput.EchoCharacter = '•'

	fields := []FormField{
		{
			Key:         "new_password",
			Title:       locale.ResetPasswordNewPassword,
			Description: locale.ResetPasswordNewPasswordDesc,
			Placeholder: "",
			Required:    true,
			Masked:      true,
			Input:       newPasswordInput,
			Value:       newPasswordInput.Value(),
		},
		{
			Key:         "confirm_password",
			Title:       locale.ResetPasswordConfirmPassword,
			Description: locale.ResetPasswordConfirmPasswordDesc,
			Placeholder: "",
			Required:    true,
			Masked:      true,
			Input:       confirmPasswordInput,
			Value:       confirmPasswordInput.Value(),
		},
	}

	h.setFormFields(fields)
	return nil
}

// setFormFields is a helper method to store fields
func (h *ResetPasswordHandler) setFormFields(fields []FormField) {
	h.fields = fields
}

// GetFormSummary returns status or error message
func (h *ResetPasswordHandler) GetFormSummary() string {
	return h.summary
}

// GetHelpContent returns help content for the form
func (h *ResetPasswordHandler) GetHelpContent() string {
	return locale.ResetPasswordHelpContent
}

// HandleSave processes password reset with form closure
func (h *ResetPasswordHandler) HandleSave() error {
	return h.processPasswordReset(true)
}

// HandleReset resets form fields
func (h *ResetPasswordHandler) HandleReset() {
	h.summary = ""
}

// OnFieldChanged is called when form field changes
func (h *ResetPasswordHandler) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// clear summary when user starts typing
	if h.summary != "" {
		h.summary = ""
	}
}

// processPasswordReset handles the password reset logic
func (h *ResetPasswordHandler) processPasswordReset(closeOnSuccess bool) error {
	// this will be called by the form screen to execute the actual operation
	// the form screen will handle field validation and call this method
	return nil
}

// GetFormFields returns current form fields (required by interface)
func (h *ResetPasswordHandler) GetFormFields() []FormField {
	return h.fields
}

// SetFormFields sets form fields (required by interface)
func (h *ResetPasswordHandler) SetFormFields(fields []FormField) {
	h.fields = fields
}

// ResetPasswordModel represents the password reset screen
type ResetPasswordModel struct {
	*BaseScreen
	*ResetPasswordHandler

	// processor integration
	processor processor.ProcessorModel

	// operation state
	operationRunning bool
	closeOnSuccess   bool
}

// NewResetPasswordModel creates a new password reset model
func NewResetPasswordModel(
	c controller.Controller, s styles.Styles, w window.Window, p processor.ProcessorModel,
) *ResetPasswordModel {
	handler := NewResetPasswordHandler(c, s, w)
	baseScreen := NewBaseScreen(c, s, w, handler, nil)

	return &ResetPasswordModel{
		BaseScreen:           baseScreen,
		ResetPasswordHandler: handler,
		processor:            p,
	}
}

// GetFormTitle returns screen title
func (m *ResetPasswordModel) GetFormTitle() string {
	return locale.ResetPasswordFormTitle
}

// GetFormDescription returns screen description
func (m *ResetPasswordModel) GetFormDescription() string {
	return locale.ResetPasswordFormDescription
}

// GetFormName returns screen name
func (m *ResetPasswordModel) GetFormName() string {
	return locale.ResetPasswordFormName
}

// GetFormOverview returns screen overview
func (m *ResetPasswordModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.ResetPasswordFormTitle))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Bold(true).Render(locale.ResetPasswordFormDescription))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(locale.ResetPasswordFormOverview))

	return strings.Join(sections, "\n")
}

// GetCurrentConfiguration returns current configuration status
func (m *ResetPasswordModel) GetCurrentConfiguration() string {
	checker := m.GetController().GetChecker()
	if !checker.PentagiRunning {
		return locale.ResetPasswordNotAvailable
	}
	return locale.ResetPasswordAvailable
}

// IsConfigured returns true if password reset is available
func (m *ResetPasswordModel) IsConfigured() bool {
	checker := m.GetController().GetChecker()
	return checker.PentagiRunning
}

// Update handles screen updates including processor messages
func (m *ResetPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case processor.ProcessorCompletionMsg:
		m.operationRunning = false
		if msg.Error != nil {
			m.summary = m.GetStyles().Error.Render(locale.ResetPasswordErrorPrefix + msg.Error.Error())
		} else {
			m.summary = m.GetStyles().Success.Render(locale.ResetPasswordSuccess)
			if m.closeOnSuccess {
				// clear form and return to previous screen
				m.HandleReset()
				return m, func() tea.Msg {
					return NavigationMsg{GoBack: true}
				}
			} else {
				// just clear form fields but stay on screen
				fields := m.GetFormFields()
				for i := range fields {
					fields[i].Value = ""
					fields[i].Input.SetValue("")
				}
				m.SetFormFields(fields)
			}
		}
		m.updateViewports()
		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorStartedMsg:
		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorOutputMsg:
		return m, m.processor.HandleMsg(msg)

	case processor.ProcessorWaitMsg:
		return m, m.processor.HandleMsg(msg)

	case tea.KeyMsg:
		// handle custom key actions first
		switch msg.String() {
		case "enter":
			if !m.operationRunning {
				return m, m.handleFormSubmission(true)
			}
			return m, nil
		case "ctrl+s":
			if !m.operationRunning {
				return m, m.handleFormSubmission(false)
			}
			return m, nil
		default:
			// delegate all other key messages to HandleFieldInput for normal typing
			if cmd := m.HandleFieldInput(msg); cmd != nil {
				return m, cmd
			}
		}
	}

	// delegate to base screen for other messages
	cmd := m.BaseScreen.Update(msg)
	return m, cmd
}

// GetFormHotKeys returns the hotkeys for this screen
func (m *ResetPasswordModel) GetFormHotKeys() []string {
	return []string{"down|up", "ctrl+s", "ctrl+h", "enter"}
}

// executePasswordReset performs the actual password reset operation
func (m *ResetPasswordModel) executePasswordReset(newPassword string, closeOnSuccess bool) tea.Cmd {
	m.operationRunning = true
	m.closeOnSuccess = closeOnSuccess
	m.summary = locale.ResetPasswordInProgress
	m.updateViewports()

	return m.processor.ResetPassword(
		context.Background(),
		processor.ProductStackPentagi,
		processor.WithPasswordValue(newPassword),
	)
}

// validatePasswords validates that passwords match and meet requirements
func (m *ResetPasswordModel) validatePasswords(newPassword, confirmPassword string) error {
	if newPassword == "" {
		return fmt.Errorf(locale.ResetPasswordErrorEmptyPassword)
	}

	if len(newPassword) < 5 {
		return fmt.Errorf(locale.ResetPasswordErrorShortPassword)
	}

	if newPassword != confirmPassword {
		return fmt.Errorf(locale.ResetPasswordErrorMismatch)
	}

	return nil
}

// handleFormSubmission processes form submission (Enter key or Ctrl+S)
func (m *ResetPasswordModel) handleFormSubmission(closeOnSuccess bool) tea.Cmd {
	if m.operationRunning {
		return nil
	}

	fields := m.GetFormFields()
	if len(fields) < 2 {
		return nil
	}

	newPassword := strings.TrimSpace(fields[0].Input.Value())
	confirmPassword := strings.TrimSpace(fields[1].Input.Value())

	if err := m.validatePasswords(newPassword, confirmPassword); err != nil {
		m.summary = m.GetStyles().Error.Render(err.Error())
		m.updateViewports()
		return nil
	}

	return m.executePasswordReset(newPassword, closeOnSuccess)
}
