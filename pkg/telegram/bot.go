package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	handler *Handler
	auth    *Auth
	stopCh  chan struct{}
}

func New(token string, allowedIDs []int64, svc FlowService) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	api.Debug = false
	log.Printf("Telegram bot authorized as @%s", api.Self.UserName)
	return &Bot{
		api:     api,
		handler: NewHandler(api, svc),
		auth:    NewAuth(allowedIDs),
		stopCh:  make(chan struct{}),
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	log.Println("Telegram bot polling for updates...")
	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			if !b.auth.IsAllowed(update.Message.From.ID) {
				log.Printf("Telegram: blocked user %d", update.Message.From.ID)
				continue
			}
			go b.handler.Handle(update.Message)
		case <-ctx.Done():
			return
		case <-b.stopCh:
			return
		}
	}
}

func (b *Bot) Stop() {
	close(b.stopCh)
	b.api.StopReceivingUpdates()
}