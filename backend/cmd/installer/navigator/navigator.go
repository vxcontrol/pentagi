package navigator

import (
	"strings"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/models"
)

type Navigator interface {
	Push(screenID models.ScreenID)
	Pop() models.ScreenID
	Current() models.ScreenID
	CanGoBack() bool
	GetStack() NavigatorStack
	String() string
}

// navigator handles screen navigation and step persistence
type navigator struct {
	stack        NavigatorStack
	stateManager state.State
}

type NavigatorStack []models.ScreenID

func (s NavigatorStack) Strings() []string {
	strings := make([]string, len(s))
	for i, v := range s {
		strings[i] = string(v)
	}
	return strings
}

func (s NavigatorStack) String() string {
	return strings.Join(s.Strings(), " -> ")
}

func NewNavigator(state state.State, checkResult checker.CheckResult) Navigator {
	logger.Log("[Nav] NEW: %s", strings.Join(state.GetStack(), " -> "))

	if !checkResult.IsReadyToContinue() {
		state.SetStack([]string{string(models.WelcomeScreen)})
		return &navigator{
			stack:        []models.ScreenID{models.WelcomeScreen},
			stateManager: state,
		}
	}

	stack := make([]models.ScreenID, 0)
	for _, screenID := range state.GetStack() {
		stack = append(stack, models.ScreenID(screenID))
	}

	return &navigator{
		stack:        stack,
		stateManager: state,
	}
}

func (n *navigator) Push(screenID models.ScreenID) {
	logger.Log("[Nav] PUSH: %s -> %s", n.Current(), screenID)
	n.stack = append(n.stack, screenID)
	n.stateManager.SetStack(n.stack.Strings())
}

func (n *navigator) Pop() models.ScreenID {
	current := n.Current()
	if len(n.stack) <= 1 {
		return current
	}

	n.stack = n.stack[:len(n.stack)-1]
	previous := n.Current()
	n.stateManager.SetStack(n.stack.Strings())
	logger.Log("[Nav] POP: %s -> %s", current, previous)
	return previous
}

func (n *navigator) Current() models.ScreenID {
	if len(n.stack) == 0 {
		return models.WelcomeScreen
	}
	return n.stack[len(n.stack)-1]
}

func (n *navigator) CanGoBack() bool {
	return len(n.stack) > 1
}

func (n *navigator) GetStack() NavigatorStack {
	return n.stack
}

func (n *navigator) String() string {
	return n.stack.String()
}
