package telegram

import (
	"fmt"
	"strings"
)

func FormatWelcome() string {
	return `*Welcome to PentAGI* 

I am your AI-powered security assistant.

Send /help to see available commands.`
}

func FormatHelp() string {
	return `*Available commands*

/flows — list your recent flows
/new <task> — create a new flow
/status <id> — check flow status
/stop <id> — stop a running flow
/help — show this message`
}

func FormatFlowList(flows []Flow) string {
	if len(flows) == 0 {
		return "No flows found. Use /new <task> to create one."
	}
	var sb strings.Builder
	sb.WriteString("*Your flows:*\n\n")
	for _, f := range flows {
		sb.WriteString(fmt.Sprintf("• `%s` — %s (%s)\n", f.ID[:8], f.Title, f.Status))
	}
	return sb.String()
}

func FormatFlowCreated(f *Flow) string {
	return fmt.Sprintf("*Flow created*\n\nID: `%s`\nTask: %s\nStatus: %s", f.ID, f.Title, f.Status)
}

func FormatFlowStatus(f *Flow) string {
	return fmt.Sprintf("*Flow status*\n\nID: `%s`\nTask: %s\nStatus: %s", f.ID, f.Title, f.Status)
}