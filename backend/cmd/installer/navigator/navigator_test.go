package navigator

import (
	"fmt"
	"reflect"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/models"
)

type mockState struct {
	stack []string
}

func (m *mockState) Exists() bool                        { return true }
func (m *mockState) Reset() error                        { return nil }
func (m *mockState) Commit() error                       { return nil }
func (m *mockState) IsDirty() bool                       { return false }
func (m *mockState) GetEulaConsent() bool                { return false }
func (m *mockState) SetEulaConsent() error               { return nil }
func (m *mockState) GetStack() []string                  { return m.stack }
func (m *mockState) SetStack(stack []string) error       { m.stack = stack; return nil }
func (m *mockState) GetVar(string) (loader.EnvVar, bool) { return loader.EnvVar{}, false }
func (m *mockState) SetVar(string, string) error         { return nil }
func (m *mockState) ResetVar(string) error               { return nil }
func (m *mockState) GetVars([]string) (map[string]loader.EnvVar, map[string]bool) {
	return nil, nil
}
func (m *mockState) SetVars(map[string]string) error      { return nil }
func (m *mockState) ResetVars([]string) error             { return nil }
func (m *mockState) GetAllVars() map[string]loader.EnvVar { return nil }
func (m *mockState) GetEnvPath() string                   { return "" }

func newMockCheckResult() checker.CheckResult {
	return checker.CheckResult{
		EnvFileExists:          true,
		EnvDirWritable:         true,
		DockerApiAccessible:    true,
		WorkerEnvApiAccessible: true,
		DockerComposeInstalled: true,
		DockerVersionOK:        true,
		DockerComposeVersionOK: true,
		SysNetworkOK:           true,
		SysCPUOK:               true,
		SysMemoryOK:            true,
		SysDiskFreeSpaceOK:     true,
	}
}

func newTestNavigator() (Navigator, *mockState) {
	mockState := &mockState{}
	nav := NewNavigator(mockState, newMockCheckResult())
	return nav, mockState
}

func TestNewNavigator(t *testing.T) {
	nav, state := newTestNavigator()

	if nav.Current() != models.WelcomeScreen {
		t.Errorf("expected WelcomeScreen, got %s", nav.Current())
	}

	if nav.CanGoBack() {
		t.Error("new navigator should not allow going back")
	}

	if len(state.stack) != 0 {
		t.Errorf("expected empty state stack, got %v", state.stack)
	}
}

func TestPushSingleScreen(t *testing.T) {
	nav, state := newTestNavigator()

	nav.Push(models.MainMenuScreen)

	if nav.Current() != models.MainMenuScreen {
		t.Errorf("expected MainMenuScreen, got %s", nav.Current())
	}

	expected := []string{string(models.MainMenuScreen)}
	if !reflect.DeepEqual(state.stack, expected) {
		t.Errorf("expected state stack %v, got %v", expected, state.stack)
	}
}

func TestPushMultipleScreens(t *testing.T) {
	nav, state := newTestNavigator()

	screens := []models.ScreenID{
		models.MainMenuScreen,
		models.LLMProvidersScreen,
		models.LLMProviderOpenAIScreen,
	}

	for _, screen := range screens {
		nav.Push(screen)
	}

	if nav.Current() != models.LLMProviderOpenAIScreen {
		t.Errorf("expected LLMProviderFormScreen, got %s", nav.Current())
	}

	expected := []string{
		string(models.MainMenuScreen),
		string(models.LLMProvidersScreen),
		string(models.LLMProviderOpenAIScreen),
	}

	if !reflect.DeepEqual(state.stack, expected) {
		t.Errorf("expected state stack %v, got %v", expected, state.stack)
	}
}

func TestCanGoBack(t *testing.T) {
	nav, _ := newTestNavigator()

	if nav.CanGoBack() {
		t.Error("empty navigator should not allow going back")
	}

	nav.Push(models.MainMenuScreen)
	if nav.CanGoBack() {
		t.Error("single screen navigator should not allow going back")
	}

	nav.Push(models.LLMProvidersScreen)
	if !nav.CanGoBack() {
		t.Error("multi-screen navigator should allow going back")
	}
}

func TestPopNormalCase(t *testing.T) {
	nav, state := newTestNavigator()

	nav.Push(models.MainMenuScreen)
	nav.Push(models.LLMProvidersScreen)
	nav.Push(models.LLMProviderOpenAIScreen)

	previous := nav.Pop()

	if previous != models.LLMProvidersScreen {
		t.Errorf("expected LLMProvidersScreen, got %s", previous)
	}

	if nav.Current() != models.LLMProvidersScreen {
		t.Errorf("expected current LLMProvidersScreen, got %s", nav.Current())
	}

	expected := []string{string(models.MainMenuScreen), string(models.LLMProvidersScreen)}
	if !reflect.DeepEqual(state.stack, expected) {
		t.Errorf("expected state stack %v, got %v", expected, state.stack)
	}
}

func TestPopEmptyStack(t *testing.T) {
	nav, _ := newTestNavigator()

	result := nav.Pop()

	if result != models.WelcomeScreen {
		t.Errorf("expected WelcomeScreen, got %s", result)
	}

	if nav.Current() != models.WelcomeScreen {
		t.Errorf("expected current WelcomeScreen, got %s", nav.Current())
	}
}

func TestPopSingleScreen(t *testing.T) {
	nav, _ := newTestNavigator()

	nav.Push(models.MainMenuScreen)
	result := nav.Pop()

	if result != models.MainMenuScreen {
		t.Errorf("expected MainMenuScreen, got %s", result)
	}

	if nav.Current() != models.MainMenuScreen {
		t.Errorf("expected current MainMenuScreen, got %s", nav.Current())
	}
}

func TestCurrentEmptyStack(t *testing.T) {
	nav, _ := newTestNavigator()

	if nav.Current() != models.WelcomeScreen {
		t.Errorf("expected WelcomeScreen for empty stack, got %s", nav.Current())
	}
}

func TestGetStack(t *testing.T) {
	nav, _ := newTestNavigator()

	screens := []models.ScreenID{
		models.MainMenuScreen,
		models.ToolsScreen,
		models.DockerFormScreen,
	}

	for _, screen := range screens {
		nav.Push(screen)
	}

	stack := nav.GetStack()
	expected := NavigatorStack{models.MainMenuScreen, models.ToolsScreen, models.DockerFormScreen}

	if !reflect.DeepEqual(stack, expected) {
		t.Errorf("expected stack %v, got %v", expected, stack)
	}
}

func TestNavigatorStackStrings(t *testing.T) {
	stack := NavigatorStack{models.WelcomeScreen, models.MainMenuScreen, models.LLMProvidersScreen}

	strings := stack.Strings()
	expected := []string{
		string(models.WelcomeScreen),
		string(models.MainMenuScreen),
		string(models.LLMProvidersScreen),
	}

	if !reflect.DeepEqual(strings, expected) {
		t.Errorf("expected strings %v, got %v", expected, strings)
	}
}

func TestNavigatorStackString(t *testing.T) {
	stack := NavigatorStack{models.WelcomeScreen, models.MainMenuScreen, models.LLMProvidersScreen}

	result := stack.String()
	expected := fmt.Sprintf("%s -> %s -> %s", models.WelcomeScreen, models.MainMenuScreen, models.LLMProvidersScreen)

	if result != expected {
		t.Errorf("expected string %q, got %q", expected, result)
	}
}

func TestNavigatorStackStringEmpty(t *testing.T) {
	stack := NavigatorStack{}

	result := stack.String()
	expected := ""

	if result != expected {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestNavigatorString(t *testing.T) {
	nav, _ := newTestNavigator()

	nav.Push(models.MainMenuScreen)
	nav.Push(models.ToolsScreen)

	result := nav.String()
	expected := string(models.MainMenuScreen) + " -> " + string(models.ToolsScreen)

	if result != expected {
		t.Errorf("expected string %q, got %q", expected, result)
	}
}

func TestComplexNavigationFlow(t *testing.T) {
	nav, state := newTestNavigator()

	// simulate typical navigation flow
	nav.Push(models.MainMenuScreen)
	nav.Push(models.LLMProvidersScreen)
	nav.Push(models.LLMProviderOpenAIScreen)

	// go back once
	nav.Pop()
	if nav.Current() != models.LLMProvidersScreen {
		t.Errorf("expected LLMProvidersScreen after pop, got %s", nav.Current())
	}

	// navigate to different branch
	nav.Push(models.ToolsScreen)
	nav.Push(models.DockerFormScreen)

	if nav.Current() != models.DockerFormScreen {
		t.Errorf("expected DockerFormScreen, got %s", nav.Current())
	}

	// verify final state
	expected := []string{
		string(models.MainMenuScreen),
		string(models.LLMProvidersScreen),
		string(models.ToolsScreen),
		string(models.DockerFormScreen),
	}

	if !reflect.DeepEqual(state.stack, expected) {
		t.Errorf("expected final state %v, got %v", expected, state.stack)
	}

	if !nav.CanGoBack() {
		t.Error("should be able to go back in complex flow")
	}
}

func TestStateIntegrationWithExistingStack(t *testing.T) {
	// test navigator initialization with existing stack
	existingStack := []string{string(models.MainMenuScreen), string(models.ToolsScreen)}
	mockState := &mockState{stack: existingStack}

	nav := NewNavigator(mockState, newMockCheckResult())

	// navigator should start with the last screen in the stack
	if nav.Current() != models.ToolsScreen {
		t.Errorf("expected ToolsScreen on new navigator, got %s", nav.Current())
	}

	if len(nav.GetStack()) != 2 {
		t.Errorf("expected 2 screens in navigator stack, got %v", nav.GetStack())
	}
}
