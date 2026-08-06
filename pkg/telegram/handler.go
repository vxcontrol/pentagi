package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FlowService interface {
	ListFlows(userID int64) ([]Flow, error)
	CreateFlow(userID int64, task string) (*Flow, error)
	GetFlowStatus(flowID string) (*Flow, error)
	StopFlow(flowID string) error
}

type Flow struct {
	ID     string
	Title  string
	Status string
}

type Handler struct {
	api *tgbotapi.BotAPI
	svc FlowService
}

func NewHandler(api *tgbotapi.BotAPI, svc FlowService) *Handler {
	return &Handler{api: api, svc: svc}
}

func (h *Handler) Handle(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.reply(msg, FormatWelcome())
	case "help":
		h.reply(msg, FormatHelp())
	case "flows":
		flows, err := h.svc.ListFlows(msg.From.ID)
		if err != nil {
			h.reply(msg, fmt.Sprintf("Error fetching flows: %v", err))
			return
		}
		h.reply(msg, FormatFlowList(flows))
	case "new":
		task := msg.CommandArguments()
		if task == "" {
			h.reply(msg, "Usage: /new <task description>")
			return
		}
		flow, err := h.svc.CreateFlow(msg.From.ID, task)
		if err != nil {
			h.reply(msg, fmt.Sprintf("Error creating flow: %v", err))
			return
		}
		h.reply(msg, FormatFlowCreated(flow))
	case "status":
		id := msg.CommandArguments()
		if id == "" {
			h.reply(msg, "Usage: /status <flow_id>")
			return
		}
		flow, err := h.svc.GetFlowStatus(id)
		if err != nil {
			h.reply(msg, fmt.Sprintf("Error: %v", err))
			return
		}
		h.reply(msg, FormatFlowStatus(flow))
	case "stop":
		id := msg.CommandArguments()
		if id == "" {
			h.reply(msg, "Usage: /stop <flow_id>")
			return
		}
		if err := h.svc.StopFlow(id); err != nil {
			h.reply(msg, fmt.Sprintf("Error: %v", err))
			return
		}
		h.reply(msg, "Flow stopped.")
	default:
		h.reply(msg, "Unknown command. Send /help for available commands.")
	}
}

func (h *Handler) reply(msg *tgbotapi.Message, text string) {
	m := tgbotapi.NewMessage(msg.Chat.ID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	h.api.Send(m)
}